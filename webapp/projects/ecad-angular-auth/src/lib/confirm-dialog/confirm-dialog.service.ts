import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { MatDialog } from '@angular/material';
import { ConfirmDialogComponent } from './confirm-dialog/confirm-dialog.component';
import { map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class ConfirmDialogService {

  constructor(
    private dialog: MatDialog
  ) { }

  confirm(message: string): Observable<boolean> {
    return this.dialog.open(ConfirmDialogComponent, {
      data: {
        message,
      }
    }).afterClosed().pipe(map((result) => {
      return result;
    }));
  }
}
