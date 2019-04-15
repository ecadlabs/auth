import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ServiceAccountEditFormComponent } from './service-account-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule, MatInputModule } from '@angular/material';

@NgModule({
  declarations: [ServiceAccountEditFormComponent],
  imports: [CommonModule, ReactiveFormsModule, MatButtonModule, MatInputModule],
  entryComponents: [ServiceAccountEditFormComponent],
  exports: [ServiceAccountEditFormComponent]
})
export class ServiceAccountEditFormModule {}
