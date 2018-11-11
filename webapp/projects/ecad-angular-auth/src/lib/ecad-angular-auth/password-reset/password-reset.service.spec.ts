import { TestBed, inject } from '@angular/core/testing';

import { PasswordResetService } from './password-reset.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('PasswordResetService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [PasswordResetService]
    });
  });

  it('should be created', inject([PasswordResetService], (service: PasswordResetService) => {
    expect(service).toBeTruthy();
  }));
});
