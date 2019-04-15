import { Component, Inject, OnInit } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { CreateUser } from '../../ecad-angular-auth-admin/interfaces/create-user.i';

@Component({
  selector: 'auth-service-account-edit-form',
  templateUrl: './service-account-edit-form.component.html',
  styleUrls: ['./service-account-edit-form.component.scss']
})
export class ServiceAccountEditFormComponent implements OnInit {
  public userForm: FormGroup;
  public error: any = {};

  constructor(
    private dialogRef: MatDialogRef<User>,
    @Inject(MAT_DIALOG_DATA)
    public dialogData: User | null,
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    private _fb: FormBuilder
  ) {}

  public get roles() {
    return this.userService.getRoles();
  }

  public get value() {
    return JSON.stringify(this.userForm.value);
  }

  ngOnInit() {
    this.userForm = this._fb.group({
      name: ['']
    });

    if (this.dialogData) {
      this.userForm.get('name').setValue(this.dialogData.name);
    }
  }

  private async createUser() {
    const createUserPayload: CreateUser = this.userForm.value;
    await this.userService.create(createUserPayload).toPromise();
  }

  private async editUser() {
    const payload = this.userForm.value;

    await this.userService
      .update(Object.assign(payload, { id: this.dialogData.id }))
      .toPromise();
  }

  async submit() {
    try {
      if (!this.dialogData) {
        await this.createUser();
      } else {
        await this.editUser();
      }
      this.error = {};
      this.dialogRef.close();
    } catch (ex) {
      if (ex.status && ex.status === 400) {
        this.error.validationError = true;
      } else if (ex.status && ex.status === 409) {
        this.error.alreadyExistsError = true;
      } else {
        this.error.serverError = true;
      }
    }
  }
}
