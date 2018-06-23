import { Component, OnInit, Inject } from '@angular/core';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material';
import { User, UsersService, CreateUser } from '../../ecad-angular-auth-admin/users/users.service';
import { FormBuilder, Form, FormGroup, Validators } from '@angular/forms';
import { MinSelection } from './user-edit-form.validators';

@Component({
  selector: 'auth-user-edit-form',
  templateUrl: './user-edit-form.component.html',
  styleUrls: ['./user-edit-form.component.scss']
})
export class UserEditFormComponent implements OnInit {

  public userForm: FormGroup;
  public error: any = {};

  constructor(
    private dialogRef: MatDialogRef<User>,
    @Inject(MAT_DIALOG_DATA)
    private dialogData: User | null,
    private userService: UsersService,
    private _fb: FormBuilder,
  ) { }

  public get roles() {
    return this.userService.getRoles();
  }

  public get value() {
    return JSON.stringify(this.userForm.value);
  }

  ngOnInit() {
    this.userForm = this._fb.group(
      {
        'email': ['', [Validators.required, Validators.pattern(/^[\w\.]+@[a-zA-Z_]+?\.[a-zA-Z]{2,3}$/)]],
        'password': ['', !this.dialogData ? [Validators.required] : []],
        'name': [''],
        'roles': [[this.userService.getRoles()[0].value], MinSelection(1)]
      }
    );

    if (this.dialogData) {
      this.userForm.get('email').setValue(this.dialogData.email);
      this.userForm.get('name').setValue(this.dialogData.name);
      this.userForm.get('roles').setValue(Object.keys(this.dialogData.roles));
    }
  }

  async submit() {
    try {
      if (!this.dialogData) {
        const createUserPayload: CreateUser = this.userForm.value;
        createUserPayload.roles = this.userForm.value.roles.reduce((prev, val) => Object.assign(prev, {[val]: true}), {});
        await this.userService.create(createUserPayload).toPromise();
      } else {
        const payload = this.userForm.value;
        const remove = this.getDeletedRole(Object.keys(this.dialogData.roles), this.userForm.value.roles);
        const added = this.getAddedRole(Object.keys(this.dialogData.roles), this.userForm.value.roles);
        await this.userService.update(Object.assign(payload, {id: this.dialogData.id}), added, remove).toPromise();
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

  private getDeletedRole(initialRoles: string[], current: string[]) {
    return initialRoles.filter((role) => !current.includes(role));
  }

  private getAddedRole(initialRoles: string[], current: string[]) {
    return current.filter((role) => !initialRoles.includes(role));
  }
}
