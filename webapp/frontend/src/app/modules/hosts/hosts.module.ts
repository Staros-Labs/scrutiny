import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { HostsComponent } from './hosts.component';
import { hostsRoutes } from './hosts.routing';

@NgModule({
    imports: [RouterModule.forChild(hostsRoutes), HostsComponent],
})
export class HostsModule {}
