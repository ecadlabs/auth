import { TestBed, inject } from '@angular/core/testing';

import { StandardLoginService } from './standard-login.service';

describe('StandardLoginServiceService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [StandardLoginService]
    });
  });

  it('should be created', inject([StandardLoginService], (service: StandardLoginService) => {
    expect(service).toBeTruthy();
  }));
});
