import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { UsersListComponent } from './users-list.component';
import { MatTableModule, MatPaginatorModule, MatSortModule, MatSnackBarModule, MatIconModule, MatDialogModule } from '@angular/material';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { PASSWORD_RESET } from '../../ecad-angular-auth/tokens';

describe('UsersListComponent', () => {
  let component: UsersListComponent;
  let fixture: ComponentFixture<UsersListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        MatTableModule,
        MatPaginatorModule,
        MatSortModule,
        NoopAnimationsModule,
        MatSnackBarModule,
        MatIconModule,
        MatDialogModule,
      ],
      declarations: [UsersListComponent],
      providers: [
        { provide: USERS_SERVICE, useValue: {} },
        { provide: PASSWORD_RESET, useValue: {} }
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(UsersListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
