import {
  Component,
  OnInit,
  Input,
  Inject,
  Output,
  EventEmitter
} from '@angular/core';
import { FormGroup, FormBuilder } from '@angular/forms';
import { Membership } from '../../ecad-angular-auth-admin/interfaces/membership.i';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { MinSelection } from '../validators';

@Component({
  selector: 'auth-member-edit-form',
  templateUrl: './member-edit-form.component.html',
  styleUrls: ['./member-edit-form.component.scss']
})
export class MemberEditFormComponent implements OnInit {
  public memberForm: FormGroup;

  @Input()
  member: Membership;

  @Output()
  memberUpdated = new EventEmitter();

  constructor(
    private _fb: FormBuilder,
    @Inject(USERS_SERVICE)
    private userService: IUsersService
  ) {}

  public get roles() {
    return this.userService.getRoles();
  }

  private getInitialRoles() {
    return Object.keys(this.member.roles);
  }

  ngOnInit() {
    this.memberForm = this._fb.group({
      roles: [Object.keys(this.member.roles), [MinSelection]]
    });
  }

  submit() {
    const { roles } = this.memberForm.value;

    const deleted = this.getDeletedRole(this.getInitialRoles(), roles);
    const added = this.getAddedRole(this.getInitialRoles(), roles);

    this.memberUpdated.next({
      roles: {
        ...added.reduce((prev, current) => ({ ...prev, [current]: true }), {}),
        ...deleted.reduce(
          (prev, current) => ({ ...prev, [current]: false }),
          {}
        )
      }
    });
  }

  private getDeletedRole(initialRoles: string[], current: string[]) {
    return initialRoles.filter(role => !current.includes(role));
  }

  private getAddedRole(initialRoles: string[], current: string[]) {
    return current.filter(role => !initialRoles.includes(role));
  }
}
