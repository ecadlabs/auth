import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule, MatSortModule, MatPaginatorModule, MatDialogModule, MatInputModule, MatButtonModule } from '@angular/material';
import { UsersListComponent } from './users-list/users-list.component';
import { UserEditFormComponent } from './user-edit-form/user-edit-form.component';

@NgModule({
  imports: [
    CommonModule,
    MatTableModule,
    MatSortModule,
    MatPaginatorModule,
    MatDialogModule,
    MatInputModule,
    MatButtonModule
  ],
  declarations: [UsersListComponent, UserEditFormComponent],
  exports: [UsersListComponent]
})
export class EcadAngularAuthAdminComponentsModule { }
