import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { UserEditFormComponent } from './user-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AUTH_CONFIG } from '../../ecad-angular-auth/tokens';
import { MAT_DIALOG_DATA, MatDialogRef, MatSelectModule, MatInputModule } from '@angular/material';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('UserEditFormComponent', () => {
  let component: UserEditFormComponent;
  let fixture: ComponentFixture<UserEditFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ReactiveFormsModule,
        MatSelectModule,
        MatInputModule,
        NoopAnimationsModule
      ],
      declarations: [UserEditFormComponent],
      providers: [
        { provide: AUTH_CONFIG, useValue: {} },
        { provide: MAT_DIALOG_DATA, useValue: null },
        {
          provide: USERS_SERVICE, useValue: {
            getRoles: () => [{ value: 'user', displayValue: 'User' }]
          }
        },
        { provide: MatDialogRef, useValue: {} }
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(UserEditFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
