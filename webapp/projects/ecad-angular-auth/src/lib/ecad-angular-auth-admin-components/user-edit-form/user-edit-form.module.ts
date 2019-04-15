import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UserEditFormComponent } from './user-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule, MatInputModule } from '@angular/material';

@NgModule({
  declarations: [UserEditFormComponent],
  imports: [CommonModule, ReactiveFormsModule, MatButtonModule, MatInputModule],
  entryComponents: [UserEditFormComponent],
  exports: [UserEditFormComponent]
})
export class UserEditFormModule {}
