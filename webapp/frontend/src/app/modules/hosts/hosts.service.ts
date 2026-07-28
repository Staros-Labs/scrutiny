import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { getBasePath } from 'app/app.routing';
import { HostActionResponse, HostSummaryResponse } from 'app/core/models/host-management-model';
import { Observable } from 'rxjs';

@Injectable({
    providedIn: 'root',
})
export class HostsService {
    private readonly http = inject(HttpClient);
    private readonly baseUrl = getBasePath() + '/api/hosts';

    getHosts(): Observable<HostSummaryResponse> {
        return this.http.get<HostSummaryResponse>(this.baseUrl);
    }

    archiveHosts(hostIds: string[]): Observable<HostActionResponse> {
        return this.http.post<HostActionResponse>(this.baseUrl + '/archive', { host_ids: hostIds });
    }

    unarchiveHosts(hostIds: string[]): Observable<HostActionResponse> {
        return this.http.post<HostActionResponse>(this.baseUrl + '/unarchive', { host_ids: hostIds });
    }

    purgeHosts(hostIds: string[]): Observable<HostActionResponse> {
        return this.http.post<HostActionResponse>(this.baseUrl + '/purge', {
            host_ids: hostIds,
            confirmation: 'PURGE',
        });
    }
}
