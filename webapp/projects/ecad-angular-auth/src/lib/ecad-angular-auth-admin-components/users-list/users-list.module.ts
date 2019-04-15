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
import { UserEditFormModule } from '../user-edit-form/user-edit-form.module';
import { ConfirmDialogModule } from '../../confirm-dialog/confirm-dialog.module';

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
    MatTooltipModule,
    UserEditFormModule,
    ConfirmDialogModule
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
