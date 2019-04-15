import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {
  MatCardModule,
  MatDividerModule,
  MatProgressBarModule
} from '@angular/material';
import { IpCreationFormModule } from '../ip-creation-form/ip-creation-form.module';
import { IpListModule } from '../ip-list/ip-list.module';
import { MembersListModule } from '../members-list/members-list.module';
import { UserLogsModule } from '../user-logs/user-logs.module';
import { UserDetailComponent } from './user-detail.component';
import { UserDetailCardModule } from '../user-detail-card/user-detail-card.module';

@NgModule({
  declarations: [UserDetailComponent],
  imports: [
    CommonModule,
    MatCardModule,
    MembersListModule,
    UserLogsModule,
    MatDividerModule,
    IpListModule,
    IpCreationFormModule,
    UserDetailCardModule,
    UserLogsModule,
    MatProgressBarModule
  ],
  exports: [UserDetailComponent]
})
export class UserDetailModule {}
