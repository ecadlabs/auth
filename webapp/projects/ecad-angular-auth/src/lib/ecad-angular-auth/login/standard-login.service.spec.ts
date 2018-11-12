import { TestBed, inject } from '@angular/core/testing';

import { StandardLoginService } from './standard-login.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AUTH_CONFIG } from '../tokens';
import { JwtHelperService } from '../jwt-helper.service';

describe('StandardLoginServiceService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        StandardLoginService,
        { provide: AUTH_CONFIG, useValue: { tokenGetter: () => { } } },
        JwtHelperService
      ]
    });
  });

  it('should be created', inject([StandardLoginService], (service: StandardLoginService) => {
    expect(service).toBeTruthy();
  }));
});
