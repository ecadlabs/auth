import { Component, Inject, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material';
import { CreateUser } from '../../ecad-angular-auth-admin/interfaces/create-user.i';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { AuthConfig } from '../../ecad-angular-auth/interfaces/auth-config.i';
import { AUTH_CONFIG } from '../../ecad-angular-auth/tokens';

@Component({
  selector: 'auth-user-edit-form',
  templateUrl: './user-edit-form.component.html',
  styleUrls: ['./user-edit-form.component.scss']
})
export class UserEditFormComponent implements OnInit {
  public userForm: FormGroup;
  public error: any = {};

  constructor(
    @Inject(AUTH_CONFIG)
    private authConfig: AuthConfig,
    private dialogRef: MatDialogRef<User>,
    @Inject(MAT_DIALOG_DATA)
    public dialogData: User | null,
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    private _fb: FormBuilder
  ) {}

  public get value() {
    return JSON.stringify(this.userForm.value);
  }

  ngOnInit() {
    this.userForm = this._fb.group({
      email: [
        '',
        [
          Validators.required,
          Validators.pattern(
            this.authConfig.emailValidationRegex || /^.+@.+\..{2,3}$/
          )
        ]
      ],
      name: ['']
    });

    if (this.dialogData) {
      this.userForm.get('email').setValue(this.dialogData.email);
      this.userForm.get('name').setValue(this.dialogData.name);
    }
  }

  private get isEmailUpdated() {
    return this.dialogData.email !== this.userForm.value.email;
  }

  async submit() {
    try {
      if (!this.dialogData) {
        const createUserPayload: CreateUser = this.userForm.value;
        createUserPayload.roles = this.userForm.value.roles.reduce(
          (prev, val) => Object.assign(prev, { [val]: true }),
          {}
        );
        await this.userService.create(createUserPayload).toPromise();
      } else {
        const payload = this.userForm.value;
        if (this.isEmailUpdated) {
          await this.userService
            .updateEmail(this.dialogData.id, this.userForm.value.email)
            .toPromise();
        }

        await this.userService
          .update(Object.assign(payload, { id: this.dialogData.id }))
          .toPromise();
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
