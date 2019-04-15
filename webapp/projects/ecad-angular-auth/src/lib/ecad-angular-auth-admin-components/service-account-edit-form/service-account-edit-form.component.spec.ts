import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ServiceAccountEditFormComponent } from './service-account-edit-form.component';
import { ReactiveFormsModule } from '@angular/forms';
import {
  MatButtonModule,
  MatInputModule,
  MAT_DIALOG_DATA,
  MatDialogRef
} from '@angular/material';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('ServiceAccountEditFormComponent', () => {
  let component: ServiceAccountEditFormComponent;
  let fixture: ComponentFixture<ServiceAccountEditFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ServiceAccountEditFormComponent],
      imports: [
        ReactiveFormsModule,
        MatButtonModule,
        MatInputModule,
        NoopAnimationsModule
      ],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: null },
        {
          provide: USERS_SERVICE,
          useValue: {
            getRoles: () => [{ value: 'user', displayValue: 'User' }]
          }
        },
        { provide: MatDialogRef, useValue: {} }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ServiceAccountEditFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
