import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MembersListComponent } from './members-list.component';
import { AuthAdminComponentsUtilsModule } from '../auth-admin-components-utils/auth-admin-components-utils.module';
import {
  MatTableModule,
  MatButtonModule,
  MatIconModule,
  MatSortModule,
  MatDialogModule,
  MatTooltipModule
} from '@angular/material';
import { MemberEditFormModule } from '../member-edit-form/member-edit-form.module';

@NgModule({
  declarations: [MembersListComponent],
  imports: [
    CommonModule,
    AuthAdminComponentsUtilsModule,
    MemberEditFormModule,
    MatDialogModule,
    MatButtonModule,
    MatSortModule,
    MatIconModule,
    MatTableModule,
    MatTooltipModule
  ],
  exports: [MembersListComponent]
})
export class MembersListModule {}
