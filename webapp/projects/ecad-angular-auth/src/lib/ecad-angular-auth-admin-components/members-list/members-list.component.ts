import {
  Component,
  Inject,
  Input,
  OnInit,
  ViewChild,
  TemplateRef
} from '@angular/core';
import { MatSort, MatDialog, MatDialogRef } from '@angular/material';
import { Subject } from 'rxjs';
import { Membership } from '../../ecad-angular-auth-admin/interfaces/membership.i';
import {
  USER_MEMBERSHIPS_FACTORY,
  USERS_SERVICE
} from '../../ecad-angular-auth-admin/tokens';
import { FilteredDatasource } from '../../filterable-datasource/filtered-datasource';
import { UserMembershipsService } from '../../ecad-angular-auth-admin/members/members.service';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';
import { switchMap, takeWhile, tap, first } from 'rxjs/operators';

@Component({
  selector: 'auth-members-list',
  templateUrl: './members-list.component.html',
  styleUrls: ['./members-list.component.scss']
})
export class MembersListComponent implements OnInit {
  @ViewChild('editMemberRoles')
  editMemberRoles: TemplateRef<any>;

  @ViewChild('deleteMember')
  deleteMember: TemplateRef<any>;

  public error = {
    serverError: false
  };

  private _userMembershipsService: UserMembershipsService;

  private dialogRef: MatDialogRef<any>;

  public displayedColumns = [
    'tenant_id',
    'status',
    'added',
    'updated',
    'roles',
    'actions'
  ];

  @Input()
  userId: string;

  public dataSource: FilteredDatasource<Membership>;
  @ViewChild(MatSort) sort: MatSort;
  private nextPage$ = new Subject<void>();
  private prevousPage$ = new Subject<void>();

  constructor(
    @Inject(USER_MEMBERSHIPS_FACTORY)
    private userMemshipsFactory: any,
    @Inject(USERS_SERVICE)
    private userService: IUsersService,
    private dialog: MatDialog
  ) {}

  ngOnInit() {
    this._userMembershipsService = this.userMemshipsFactory(this.userId);
    this.dataSource = new FilteredDatasource(
      this._userMembershipsService,
      this.sort.sortChange,
      this.nextPage$,
      this.prevousPage$
    );
  }

  archiveMember({ tenant_id, user_id }: Membership) {
    this.dialog
      .open(this.deleteMember)
      .afterClosed()
      .pipe(
        takeWhile(x => x),
        switchMap(() => {
          return this.userService.archiveMembership(user_id, tenant_id);
        }),
        tap(() => this.dataSource.refresh()),
        first()
      )
      .subscribe();
  }

  updateMember(member: Membership) {
    this.dialogRef = this.dialog.open(this.editMemberRoles, {
      data: { member }
    });
  }

  async submitUpdateMember(member: Membership, { roles }) {
    if (this.dialogRef) {
      try {
        await this.userService
          .updateMembership({
            tenantId: member.tenant_id,
            userId: member.user_id,
            roles
          })
          .toPromise();

        this.dialogRef.close();
        this.dataSource.refresh();
      } catch (ex) {
        this.error.serverError = true;
      }
    }
  }
}
