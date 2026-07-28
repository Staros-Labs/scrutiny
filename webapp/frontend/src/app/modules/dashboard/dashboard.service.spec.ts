import { HttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { DashboardService } from './dashboard.service';
import { of } from 'rxjs';
import { summary } from 'app/data/mock/summary/data';
import { temp_history } from 'app/data/mock/summary/temp_history';
import { filesystem_summary } from 'app/data/mock/summary/filesystem_summary';
import { DeviceSummaryModel } from 'app/core/models/device-summary-model';
import { SmartTemperatureModel } from 'app/core/models/measurements/smart-temperature-model';

describe('DashboardService', () => {
    let service: DashboardService;
    let httpClientSpy: jasmine.SpyObj<HttpClient>;

    beforeEach(() => {
        httpClientSpy = jasmine.createSpyObj('HttpClient', ['get']);
        TestBed.configureTestingModule({
            providers: [DashboardService, { provide: HttpClient, useValue: httpClientSpy }],
        });
        service = TestBed.inject(DashboardService);
    });

    it('should unwrap and return getSummaryData() (HttpClient called once)', (done: DoneFn) => {
        httpClientSpy.get.and.returnValue(of(summary));

        service.getSummaryData().subscribe((value) => {
            expect(value).toBe(summary.data.summary as { [key: string]: DeviceSummaryModel });
            done();
        });
        expect(httpClientSpy.get.calls.count()).withContext('one call').toBe(1);
    });

    it('should request and publish paginated summary data', (done: DoneFn) => {
        const page = {
            success: true,
            errors: [],
            data: {
                summary: summary.data.summary,
                pagination: {
                    page: 2,
                    page_size: 50,
                    total_items: 75,
                    total_pages: 2,
                    attention_count: 4,
                },
            },
        };
        httpClientSpy.get.and.returnValue(of(page));

        service.getSummaryPage({ page: 2, pageSize: 50, archived: true, sort: 'title_asc', display: 'label' }).subscribe((value) => {
            expect(value).toEqual(page.data);
            expect(httpClientSpy.get).toHaveBeenCalledWith(jasmine.stringMatching(/\/api\/summary$/), {
                params: {
                    page: '2',
                    archived: 'true',
                    page_size: '50',
                    sort: 'title_asc',
                    display: 'label',
                },
            });
            done();
        });
    });

    it('should unwrap and return getSummaryTempData() (HttpClient called once)', (done: DoneFn) => {
        httpClientSpy.get.and.returnValue(of(temp_history));

        service.getSummaryTempData('weekly', ['device-1', 'device-2']).subscribe((value) => {
            expect(value).toBe(temp_history.data.temp_history as { [key: string]: SmartTemperatureModel[] });
            const options = httpClientSpy.get.calls.mostRecent().args[1] as any;
            expect(options.params.get('duration_key')).toBe('weekly');
            expect(options.params.getAll('device_id')).toEqual(['device-1', 'device-2']);
            done();
        });
        expect(httpClientSpy.get.calls.count()).withContext('one call').toBe(1);
    });

    it('should load temperature device options independently from dashboard pages', (done: DoneFn) => {
        const response = {
            success: true,
            data: {
                devices: [{ device_id: 'device-1', host_id: 'host-1', label: '', device_name: 'sda', model_name: 'Model', serial_number: 'Serial' }],
            },
        };
        httpClientSpy.get.and.returnValue(of(response));

        service.getTemperatureDeviceOptions().subscribe((value) => {
            expect(value).toEqual(response.data.devices);
            expect(httpClientSpy.get).toHaveBeenCalledWith(jasmine.stringMatching(/\/api\/summary\/temp\/devices$/));
            done();
        });
    });

    it('should unwrap and return getFilesystemSummaryData() (HttpClient called once)', (done: DoneFn) => {
        httpClientSpy.get.and.returnValue(of(filesystem_summary));

        service.getFilesystemSummaryData().subscribe((value) => {
            expect(value).toEqual(filesystem_summary.data);
            done();
        });
        expect(httpClientSpy.get.calls.count()).withContext('one call').toBe(1);
    });
});
