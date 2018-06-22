import { Component, OnInit, Inject } from '@angular/core';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material';
import { User } from '../../ecad-angular-auth-admin/users/users.service';

@Component({
  selector: 'auth-user-edit-form',
  templateUrl: './user-edit-form.component.html',
  styleUrls: ['./user-edit-form.component.scss']
})
export class UserEditFormComponent implements OnInit {

  constructor(
    private dialogRef: MatDialogRef<User>,
    @Inject(MAT_DIALOG_DATA)
    private dalogData: User | null
  ) { }

  ngOnInit() {
  }

}
