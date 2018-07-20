import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  MatTableModule,
  MatSortModule,
  MatPaginatorModule,
  MatDialogModule,
  MatInputModule,
  MatButtonModule,
  MatSelectModule,
  MatIconModule,
  MatSnackBarModule,
  MatGridListModule,
  MatDividerModule,
  MatCardModule
} from '@angular/material';
import { UsersListComponent } from './users-list/users-list.component';
import { UserEditFormComponent } from './user-edit-form/user-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import { ConfirmDialogModule } from '../confirm-dialog/confirm-dialog.module';
import { UserDetailComponent } from './user-detail/user-detail.component';
import { UserLogsComponent } from './user-logs/user-logs.component';

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
    MatDividerModule,
    MatCardModule
  ],
  declarations: [UsersListComponent, UserEditFormComponent, UserDetailComponent, UserLogsComponent],
  entryComponents: [UserEditFormComponent],
  exports: [UsersListComponent, UserDetailComponent, UserLogsComponent]
})
export class EcadAngularAuthAdminComponentsModule { }
