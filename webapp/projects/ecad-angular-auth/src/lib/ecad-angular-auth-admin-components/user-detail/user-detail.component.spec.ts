import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { UserDetailComponent } from './user-detail.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {
  USERS_SERVICE,
  USER_MEMBERSHIPS_FACTORY,
  USER_LOG_SERVICE
} from '../../ecad-angular-auth-admin/tokens';
import { of } from 'rxjs';
import { UserLogsModule } from '../user-logs/user-logs.module';
import { IpCreationFormModule } from '../ip-creation-form/ip-creation-form.module';
import { UserDetailCardModule } from '../user-detail-card/user-detail-card.module';
import { IpListModule } from '../ip-list/ip-list.module';
import { MembersListModule } from '../members-list/members-list.module';
import {
  MatCardModule,
  MatDividerModule,
  MatProgressBarModule
} from '@angular/material';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('UserDetailComponent', () => {
  let component: UserDetailComponent;
  let fixture: ComponentFixture<UserDetailComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [UserDetailComponent],
      imports: [
        MatCardModule,
        MembersListModule,
        UserLogsModule,
        MatDividerModule,
        IpListModule,
        IpCreationFormModule,
        UserDetailCardModule,
        MatProgressBarModule,
        NoopAnimationsModule
      ],
      providers: [
        {
          provide: USER_MEMBERSHIPS_FACTORY,
          useValue: () => new class Service {}()
        },
        {
          provide: USERS_SERVICE,
          useValue: {
            find: () => of({ roles: { user: true } }),
            getRoles: () => []
          }
        },
        { provide: USER_LOG_SERVICE, useValue: {} }
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(UserDetailComponent);
    component = fixture.componentInstance;
    component.userId = 'AN_ID';
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
