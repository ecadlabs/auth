import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ServiceAccountListComponent } from './service-account-list.component';
import {
  MatIconModule,
  MatTableModule,
  MatPaginatorModule,
  MatSortModule,
  MatButtonModule,
  MatTooltipModule
} from '@angular/material';

@NgModule({
  declarations: [ServiceAccountListComponent],
  imports: [
    CommonModule,
    MatIconModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatButtonModule,
    MatTooltipModule
  ],
  exports: [ServiceAccountListComponent]
})
export class ServiceAccountListModule {}
