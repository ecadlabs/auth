import { Component, OnInit, Inject } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material';

@Component({
  selector: 'auth-confirm-dialog',
  templateUrl: './confirm-dialog.component.html',
  styleUrls: ['./confirm-dialog.component.scss']
})
export class ConfirmDialogComponent implements OnInit {

  public get message() {
    return this.data.message;
  }

  constructor(
    private dialog: MatDialogRef<any>,
    @Inject(MAT_DIALOG_DATA)
    public data: any,
  ) { }

  ngOnInit() {
  }

  confirm() {
    this.dialog.close(true);
  }

  cancel() {
    this.dialog.close(false);
  }

}
