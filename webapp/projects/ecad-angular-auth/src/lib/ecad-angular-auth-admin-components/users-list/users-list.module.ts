import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { tmpl } from './reset-password.svg';
import {
  MatIconModule,
  MatTableModule,
  MatPaginatorModule,
  MatSortModule,
  MatButtonModule,
  MatTooltipModule,
  MatIconRegistry
} from '@angular/material';
import { UsersListComponent } from './users-list.component';
import { DomSanitizer } from '@angular/platform-browser';

@NgModule({
  declarations: [UsersListComponent],
  exports: [UsersListComponent],
  imports: [
    CommonModule,
    MatIconModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatButtonModule,
    MatTooltipModule
  ]
})
export class UsersListModule {
  constructor(matRegistry: MatIconRegistry, sanitrizer: DomSanitizer) {
    matRegistry.addSvgIconLiteral(
      'reset-password',
      sanitrizer.bypassSecurityTrustHtml(tmpl)
    );
  }
}
