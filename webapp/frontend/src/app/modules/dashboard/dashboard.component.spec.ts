import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { AppConfig } from 'app/core/config/app.config';
import { ScrutinyConfigService } from 'app/core/config/scrutiny-config.service';
import { MDADMService } from 'app/modules/mdadm/mdadm.service';
import { DashboardComponent } from './dashboard.component';
import { DashboardService } from './dashboard.service';
import { TemperatureSelection } from './temperature-selection';

describe('DashboardComponent temperature chart', () => {
    let component: DashboardComponent;
    let fixture: ComponentFixture<DashboardComponent>;
    let dashboardService: jasmine.SpyObj<DashboardService>;

    beforeEach(() => {
        const config: AppConfig = {
            dashboard_columns: 2,
            dashboard_density: 'comfortable',
            line_stroke: 'smooth',
            temperature_unit: 'celsius',
            theme: 'light',
            time_format: '24',
        };
        const temperatureSelection = new TemperatureSelection();

        dashboardService = jasmine.createSpyObj('DashboardService', ['getTemperatureDeviceOptions', 'getSummaryTempData', 'getFilesystemSummaryData'], {
            pageData$: of({
                summary: {},
                pagination: {
                    page: 1,
                    page_size: 25,
                    total_items: 0,
                    total_pages: 0,
                    attention_count: 0,
                },
            }),
            temperatureSelection,
        });
        dashboardService.getTemperatureDeviceOptions.and.returnValue(
            of([
                {
                    device_id: 'drive-1',
                    host_id: 'host-1',
                    label: '',
                    device_name: 'sda',
                    model_name: 'Test Drive',
                    serial_number: 'serial-1',
                },
            ])
        );
        dashboardService.getSummaryTempData.and.returnValue(
            of({
                'drive-1': [{ date: '2026-08-03T12:00:00Z', temp: 64 }],
            })
        );
        dashboardService.getFilesystemSummaryData.and.returnValue(of({ filesystems: {}, hosts: {} }));

        TestBed.configureTestingModule({
            imports: [DashboardComponent],
            providers: [
                { provide: DashboardService, useValue: dashboardService },
                { provide: MDADMService, useValue: { getSummaryData: () => of([]) } },
                { provide: ScrutinyConfigService, useValue: { config$: of(config) } },
                { provide: MatDialog, useValue: { open: jasmine.createSpy('open') } },
                {
                    provide: Router,
                    useValue: {
                        navigate: jasmine.createSpy('navigate'),
                        onSameUrlNavigation: 'ignore',
                        routeReuseStrategy: { shouldReuseRoute: () => true },
                        url: '/web/dashboard',
                    },
                },
            ],
        });

        fixture = TestBed.createComponent(DashboardComponent);
        component = fixture.componentInstance;
        component.ngOnInit();
    });

    afterEach(() => {
        component.ngOnDestroy();
        fixture.destroy();
    });

    it('binds loaded temperature data before the lazy chart instance exists', () => {
        component.toggleTemperatureDevice('drive-1');

        expect(component.temperatureOptions.series).toEqual([
            {
                name: 'host-1: /dev/sda - Test Drive',
                data: [{ x: new Date('2026-08-03T12:00:00Z'), y: 64 }],
            },
        ]);
    });
});
