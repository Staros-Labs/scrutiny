import { Injectable, inject } from '@angular/core';
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { DashboardService } from 'app/modules/dashboard/dashboard.service';
import { DeviceSummaryPage } from 'app/core/models/device-summary-response-wrapper';

@Injectable({
    providedIn: 'root',
})
export class DashboardResolver {
    private readonly _dashboardService = inject(DashboardService);

    // -----------------------------------------------------------------------------------------------------
    // @ Public methods
    // -----------------------------------------------------------------------------------------------------

    /**
     * Resolver
     *
     * @param route
     * @param state
     */
    resolve(_route: ActivatedRouteSnapshot, _state: RouterStateSnapshot): Observable<DeviceSummaryPage> {
        return this._dashboardService.getSummaryPage({ page: 1 });
    }
}
