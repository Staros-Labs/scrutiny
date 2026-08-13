import { HttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { HostsService } from './hosts.service';

describe('HostsService', () => {
    let service: HostsService;
    let httpClient: jasmine.SpyObj<HttpClient>;

    beforeEach(() => {
        httpClient = jasmine.createSpyObj('HttpClient', ['get', 'post']);
        TestBed.configureTestingModule({
            providers: [HostsService, { provide: HttpClient, useValue: httpClient }],
        });
        service = TestBed.inject(HostsService);
    });

    it('loads SMART host inventory', () => {
        const response = {
            success: true,
            data: [{ host_id: 'alpha', active_devices: 2, archived_devices: 1, total_devices: 3 }],
        };
        httpClient.get.and.returnValue(of(response));

        service.getHosts().subscribe((value) => expect(value).toEqual(response));

        expect(httpClient.get).toHaveBeenCalledWith(jasmine.stringMatching(/\/api\/hosts$/));
    });

    it('requires server-side PURGE confirmation for selected hosts', () => {
        const response = { success: true, data: [] };
        httpClient.post.and.returnValue(of(response));

        service.purgeHosts(['alpha', 'beta']).subscribe((value) => expect(value).toEqual(response));

        expect(httpClient.post).toHaveBeenCalledWith(jasmine.stringMatching(/\/api\/hosts\/purge$/), {
            host_ids: ['alpha', 'beta'],
            confirmation: 'PURGE',
        });
    });
});
