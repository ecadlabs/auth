import { EcadAngularAuthAdminModule } from './ecad-angular-auth-admin.module';

describe('EcadAngularAuthAdminModule', () => {
  let ecadAngularAuthAdminModule: EcadAngularAuthAdminModule;

  beforeEach(() => {
    ecadAngularAuthAdminModule = new EcadAngularAuthAdminModule();
  });

  it('should create an instance', () => {
    expect(ecadAngularAuthAdminModule).toBeTruthy();
  });
});
