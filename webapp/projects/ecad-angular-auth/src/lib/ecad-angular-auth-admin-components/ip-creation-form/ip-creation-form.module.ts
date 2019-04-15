import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { IpCreationFormComponent } from './ip-creation-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import { MatInputModule, MatButtonModule } from '@angular/material';

@NgModule({
  declarations: [IpCreationFormComponent],
  imports: [ReactiveFormsModule, MatInputModule, MatButtonModule, CommonModule],
  exports: [IpCreationFormComponent]
})
export class IpCreationFormModule {}
