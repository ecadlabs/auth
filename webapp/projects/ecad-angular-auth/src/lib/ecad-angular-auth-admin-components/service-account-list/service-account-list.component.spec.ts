import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ServiceAccountListComponent } from './service-account-list.component';
import {
  MatTableModule,
  MatPaginatorModule,
  MatSortModule,
  MatSnackBarModule,
  MatIconModule,
  MatDialogModule
} from '@angular/material';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { USERS_SERVICE } from '../../ecad-angular-auth-admin/tokens';
import { PASSWORD_RESET } from '../../ecad-angular-auth/tokens';

describe('ServiceAccountListComponent', () => {
  let component: ServiceAccountListComponent;
  let fixture: ComponentFixture<ServiceAccountListComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        MatTableModule,
        MatPaginatorModule,
        MatSortModule,
        NoopAnimationsModule,
        MatSnackBarModule,
        MatIconModule,
        MatDialogModule
      ],
      providers: [
        { provide: USERS_SERVICE, useValue: {} },
        { provide: PASSWORD_RESET, useValue: {} }
      ],
      declarations: [ServiceAccountListComponent]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ServiceAccountListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
