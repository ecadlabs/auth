import { NgModule, ModuleWithProviders } from '@angular/core';
import { CommonModule } from '@angular/common';

export const authAdminConfig = 'AUTH_ADMIN_CONFIG';

export interface AuthAdminConfig {
  roles: {
    value: string;
    displayValue: string;
  }[];
  apiEndpoint: string;
}

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: []
})
export class EcadAngularAuthAdminModule {
  public static forRoot(config: AuthAdminConfig): ModuleWithProviders {
    return {
     ngModule: EcadAngularAuthAdminModule,
     providers: [
         { provide: authAdminConfig, useValue: config },
     ]
   };
 }
}
