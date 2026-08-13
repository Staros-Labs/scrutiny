import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { of } from 'rxjs';
import { HostsComponent } from './hosts.component';
import { HostsService } from './hosts.service';

describe('HostsComponent', () => {
    let component: HostsComponent;
    let fixture: ComponentFixture<HostsComponent>;
    let hostsService: jasmine.SpyObj<HostsService>;
    let dialog: jasmine.SpyObj<MatDialog>;

    beforeEach(async () => {
        hostsService = jasmine.createSpyObj('HostsService', ['getHosts', 'archiveHosts', 'unarchiveHosts', 'purgeHosts']);
        dialog = jasmine.createSpyObj('MatDialog', ['open']);
        hostsService.getHosts.and.returnValue(
            of({
                success: true,
                data: [
                    { host_id: 'alpha', active_devices: 2, archived_devices: 1, total_devices: 3 },
                    { host_id: 'beta', active_devices: 1, archived_devices: 0, total_devices: 1 },
                ],
            })
        );

        await TestBed.configureTestingModule({
            imports: [NoopAnimationsModule, HostsComponent],
            providers: [
                { provide: HostsService, useValue: hostsService },
                { provide: MatDialog, useValue: dialog },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(HostsComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('filters hosts and keeps selection independent from search', () => {
        component.toggleHost('alpha', true);
        component.setSearch('beta');

        expect(component.filteredHosts.map((host) => host.host_id)).toEqual(['beta']);
        expect(component.selectedHostIds.has('alpha')).toBeTrue();
    });

    it('archives all selected host devices and refreshes counts', () => {
        hostsService.archiveHosts.and.returnValue(
            of({
                success: true,
                data: [{ host_id: 'alpha', success: true, device_count: 3 }],
            })
        );
        component.toggleHost('alpha', true);

        component.archiveSelected();

        expect(hostsService.archiveHosts).toHaveBeenCalledWith(['alpha']);
        expect(hostsService.getHosts).toHaveBeenCalledTimes(2);
    });

    it('purges only after typed confirmation dialog succeeds', () => {
        hostsService.purgeHosts.and.returnValue(
            of({
                success: false,
                data: [{ host_id: 'alpha', success: false, device_count: 3, error: 'storage unavailable' }],
            })
        );
        dialog.open.and.returnValue({ afterClosed: () => of(true) } as any);
        component.toggleHost('alpha', true);

        component.purgeSelected();

        expect(hostsService.purgeHosts).toHaveBeenCalledWith(['alpha']);
        expect(component.selectedHostIds.has('alpha')).toBeTrue();
        expect(component.lastAction).toBe('purge');
    });
});
