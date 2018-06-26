import { NgModule, ModuleWithProviders } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthAdminConfig } from './interfaces';
import { authAdminConfig } from './tokens';

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
