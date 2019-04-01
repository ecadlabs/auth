import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatSelectModule, MatButtonModule } from '@angular/material';
import { MemberEditFormComponent } from './member-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';

@NgModule({
  declarations: [MemberEditFormComponent],
  imports: [
    CommonModule,
    MatSelectModule,
    ReactiveFormsModule,
    MatButtonModule
  ],
  exports: [MemberEditFormComponent]
})
export class MemberEditFormModule {}
