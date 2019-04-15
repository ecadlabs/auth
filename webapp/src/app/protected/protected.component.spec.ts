import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ProtectedComponent } from './protected.component';
import { ecadAngularAuth } from 'src/testing/fixture';
import { MatSnackBarModule } from '@angular/material';

describe('ProtectedComponent', () => {
  let component: ProtectedComponent;
  let fixture: ComponentFixture<ProtectedComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [...ecadAngularAuth, MatSnackBarModule],
      declarations: [ProtectedComponent]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProtectedComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
