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
  MatIconModule
} from '@angular/material';
import { UsersListComponent } from './users-list/users-list.component';
import { UserEditFormComponent } from './user-edit-form/user-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';

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
    MatIconModule
  ],
  declarations: [UsersListComponent, UserEditFormComponent],
  entryComponents: [UserEditFormComponent],
  exports: [UsersListComponent]
})
export class EcadAngularAuthAdminComponentsModule { }
