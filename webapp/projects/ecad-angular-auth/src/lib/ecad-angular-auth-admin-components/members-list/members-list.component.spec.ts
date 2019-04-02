import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { MembersListComponent } from './members-list.component';
import {
  MatDialogModule,
  MatButtonModule,
  MatSortModule,
  MatIconModule,
  MatTableModule,
  MatTooltipModule
} from '@angular/material';
import { AuthAdminComponentsUtilsModule } from '../auth-admin-components-utils/auth-admin-components-utils.module';
import { Component } from '@angular/core';
import {
  USER_MEMBERSHIPS_FACTORY,
  USERS_SERVICE
} from '../../ecad-angular-auth-admin/tokens';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

@Component({
  selector: 'auth-member-edit-form',
  // tslint:disable-next-line: use-input-property-decorator
  inputs: ['member'],
  template: ''
})
// tslint:disable-next-line: component-class-suffix
export class MockMemberEditForm {}

describe('MembersListComponent', () => {
  let component: MembersListComponent;
  let fixture: ComponentFixture<MembersListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [MembersListComponent, MockMemberEditForm],
      imports: [
        AuthAdminComponentsUtilsModule,
        MatDialogModule,
        MatButtonModule,
        MatSortModule,
        MatIconModule,
        MatTableModule,
        MatTooltipModule,
        NoopAnimationsModule
      ],
      providers: [
        {
          provide: USER_MEMBERSHIPS_FACTORY,
          useValue: () => new class Service {}()
        },
        {
          provide: USERS_SERVICE,
          useValue: {}
        }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(MembersListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
