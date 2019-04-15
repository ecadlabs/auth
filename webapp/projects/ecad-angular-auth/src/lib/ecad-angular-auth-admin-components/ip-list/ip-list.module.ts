import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { IpListComponent } from './ip-list.component';
import {
  MatButtonModule,
  MatSortModule,
  MatIconModule,
  MatTableModule,
  MatTooltipModule
} from '@angular/material';

@NgModule({
  declarations: [IpListComponent],
  imports: [
    CommonModule,
    MatButtonModule,
    MatSortModule,
    MatIconModule,
    MatTableModule,
    MatTooltipModule
  ],
  exports: [IpListComponent]
})
export class IpListModule {}
