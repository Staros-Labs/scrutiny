import { ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit, ViewEncapsulation, inject } from '@angular/core';
import { MatButton } from '@angular/material/button';
import { MatCheckbox } from '@angular/material/checkbox';
import { MatDialog } from '@angular/material/dialog';
import { MatFormField, MatLabel, MatPrefix } from '@angular/material/form-field';
import { MatIcon } from '@angular/material/icon';
import { MatInput } from '@angular/material/input';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { HostActionResponse, HostActionResultModel, HostSummaryModel } from 'app/core/models/host-management-model';
import { Observable } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { HostPurgeDialogComponent } from './host-purge-dialog.component';
import { HostsService } from './hosts.service';

@Component({
    selector: 'hosts',
    templateUrl: './hosts.component.html',
    styleUrls: ['./hosts.component.scss'],
    encapsulation: ViewEncapsulation.None,
    changeDetection: ChangeDetectionStrategy.OnPush,
    imports: [MatButton, MatCheckbox, MatFormField, MatLabel, MatPrefix, MatIcon, MatInput, MatProgressSpinner],
})
export class HostsComponent implements OnInit {
    private readonly hostsService = inject(HostsService);
    private readonly dialog = inject(MatDialog);
    private readonly changeDetectorRef = inject(ChangeDetectorRef);

    hosts: HostSummaryModel[] = [];
    selectedHostIds = new Set<string>();
    search = '';
    loading = true;
    actionInProgress = false;
    actionResults: HostActionResultModel[] = [];
    lastAction: 'archive' | 'unarchive' | 'purge' | null = null;
    loadError = '';

    get filteredHosts(): HostSummaryModel[] {
        const search = this.search.trim().toLowerCase();
        if (!search) {
            return this.hosts;
        }
        return this.hosts.filter((host) => host.host_id.toLowerCase().includes(search));
    }

    get selectedCount(): number {
        return this.selectedHostIds.size;
    }

    get allFilteredSelected(): boolean {
        return this.filteredHosts.length > 0 && this.filteredHosts.every((host) => this.selectedHostIds.has(host.host_id));
    }

    ngOnInit(): void {
        this.loadHosts();
    }

    loadHosts(): void {
        this.loading = true;
        this.loadError = '';
        this.hostsService
            .getHosts()
            .pipe(
                finalize(() => {
                    this.loading = false;
                    this.changeDetectorRef.markForCheck();
                })
            )
            .subscribe({
                next: (response) => {
                    this.hosts = response.data;
                    const availableHostIds = new Set(this.hosts.map((host) => host.host_id));
                    this.selectedHostIds = new Set([...this.selectedHostIds].filter((hostId) => availableHostIds.has(hostId)));
                },
                error: () => {
                    this.loadError = 'Could not load SMART hosts.';
                },
            });
    }

    setSearch(value: string): void {
        this.search = value;
    }

    toggleHost(hostId: string, checked: boolean): void {
        if (checked) {
            this.selectedHostIds.add(hostId);
        } else {
            this.selectedHostIds.delete(hostId);
        }
    }

    toggleFilteredHosts(checked: boolean): void {
        for (const host of this.filteredHosts) {
            if (checked) {
                this.selectedHostIds.add(host.host_id);
            } else {
                this.selectedHostIds.delete(host.host_id);
            }
        }
    }

    archiveSelected(): void {
        this.runAction('archive', (hostIds) => this.hostsService.archiveHosts(hostIds));
    }

    unarchiveSelected(): void {
        this.runAction('unarchive', (hostIds) => this.hostsService.unarchiveHosts(hostIds));
    }

    purgeSelected(): void {
        const hostIds = [...this.selectedHostIds];
        if (hostIds.length === 0) {
            return;
        }
        this.dialog
            .open(HostPurgeDialogComponent, {
                width: '520px',
                maxWidth: '95vw',
                data: { hostIds },
            })
            .afterClosed()
            .subscribe((confirmed) => {
                if (confirmed) {
                    this.runAction('purge', (selectedHostIds) => this.hostsService.purgeHosts(selectedHostIds));
                }
            });
    }

    retryFailed(): void {
        const failedHostIds = this.actionResults.filter((result) => !result.success).map((result) => result.host_id);
        this.selectedHostIds = new Set(failedHostIds);
        this.purgeSelected();
    }

    private runAction(actionName: 'archive' | 'unarchive' | 'purge', action: (hostIds: string[]) => Observable<HostActionResponse>): void {
        const hostIds = [...this.selectedHostIds];
        if (hostIds.length === 0 || this.actionInProgress) {
            return;
        }
        this.actionInProgress = true;
        this.lastAction = actionName;
        this.actionResults = [];
        action(hostIds)
            .pipe(
                finalize(() => {
                    this.actionInProgress = false;
                    this.changeDetectorRef.markForCheck();
                })
            )
            .subscribe({
                next: (response) => {
                    this.actionResults = response.data;
                    this.selectedHostIds = new Set(response.data.filter((result) => !result.success).map((result) => result.host_id));
                    this.loadHosts();
                },
                error: () => {
                    this.actionResults = hostIds.map((hostId) => ({
                        host_id: hostId,
                        success: false,
                        device_count: 0,
                        error: 'Request failed before host result was returned.',
                    }));
                },
            });
    }
}
