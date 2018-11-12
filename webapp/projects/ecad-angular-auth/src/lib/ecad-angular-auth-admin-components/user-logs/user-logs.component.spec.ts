import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { UserLogsComponent } from './user-logs.component';
import { MatTableModule, MatPaginatorModule, MatSortModule } from '@angular/material';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { USER_LOG_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('UserLogsComponent', () => {
  let component: UserLogsComponent;
  let fixture: ComponentFixture<UserLogsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [MatTableModule, MatPaginatorModule, MatSortModule, NoopAnimationsModule],
      declarations: [UserLogsComponent],
      providers: [
        { provide: USERS_SERVICE, useValue: {} },
        { provide: USER_LOG_SERVICE, useValue: {} }
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(UserLogsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
