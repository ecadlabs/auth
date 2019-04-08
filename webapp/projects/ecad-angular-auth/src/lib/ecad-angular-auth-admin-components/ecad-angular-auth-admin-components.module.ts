import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import {
  MatButtonModule,
  MatCardModule,
  MatDialogModule,
  MatDividerModule,
  MatIconModule,
  MatInputModule,
  MatPaginatorModule,
  MatSelectModule,
  MatSnackBarModule,
  MatSortModule,
  MatTableModule,
  MatTooltipModule
} from '@angular/material';
import { ConfirmDialogModule } from '../confirm-dialog/confirm-dialog.module';
import { MembersListModule } from './members-list/members-list.module';
import { UserDetailComponent } from './user-detail/user-detail.component';
import { UserEditFormComponent } from './user-edit-form/user-edit-form.component';
import { UserLogsComponent } from './user-logs/user-logs.component';
import { UsersListModule } from './users-list/users-list.module';
import { ServiceAccountListModule } from './service-account-list/service-account-list.module';
@NgModule({
  imports: [
    CommonModule,
    MatTableModule,
    MatSortModule,
    MatPaginatorModule,
    MatDialogModule,
    MatInputModule,
    MatButtonModule,
    ReactiveFormsModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatSnackBarModule,
    ConfirmDialogModule,
    MatTooltipModule,
    MatDividerModule,
    MatCardModule,
    MembersListModule,
    UsersListModule,
    ServiceAccountListModule
  ],
  declarations: [UserEditFormComponent, UserDetailComponent, UserLogsComponent],
  entryComponents: [UserEditFormComponent],
  exports: [
    UsersListModule,
    ServiceAccountListModule,
    UserDetailComponent,
    UserLogsComponent
  ]
})
export class EcadAngularAuthAdminComponentsModule {}
