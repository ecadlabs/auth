import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UserLogsComponent } from './user-logs.component';
import {
  MatButtonModule,
  MatSortModule,
  MatIconModule,
  MatTableModule,
  MatTooltipModule,
  MatPaginatorModule
} from '@angular/material';

@NgModule({
  declarations: [UserLogsComponent],
  imports: [
    CommonModule,
    MatButtonModule,
    MatSortModule,
    MatIconModule,
    MatTableModule,
    MatTooltipModule,
    MatPaginatorModule
  ],
  exports: [UserLogsComponent]
})
export class UserLogsModule {}
