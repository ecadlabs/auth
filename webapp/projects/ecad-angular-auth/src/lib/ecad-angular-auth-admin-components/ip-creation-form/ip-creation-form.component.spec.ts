import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { IpCreationFormComponent } from './ip-creation-form.component';
import { MatInputModule, MatButtonModule } from '@angular/material';
import { ReactiveFormsModule } from '@angular/forms';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

describe('IpCreationFormComponent', () => {
  let component: IpCreationFormComponent;
  let fixture: ComponentFixture<IpCreationFormComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [IpCreationFormComponent],
      imports: [
        MatInputModule,
        MatButtonModule,
        ReactiveFormsModule,
        NoopAnimationsModule
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(IpCreationFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
