import { RolesPipe } from './roles.pipe';
import { IUsersService } from '../../ecad-angular-auth-admin/interfaces/user-service.i';

describe('RolesPipe', () => {
  it('create an instance', () => {
    const fakeUserService = {
      getRoles: () => [{ displayValue: 'Test', value: 'test' }]
    };
    const pipe = new RolesPipe(fakeUserService as IUsersService);
    expect(pipe).toBeTruthy();
  });
});
