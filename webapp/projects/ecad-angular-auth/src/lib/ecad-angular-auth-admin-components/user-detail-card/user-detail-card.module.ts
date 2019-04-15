import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UserDetailCardComponent } from './user-detail-card.component';
import { MatCardModule, MatDividerModule } from '@angular/material';

@NgModule({
  declarations: [UserDetailCardComponent],
  imports: [CommonModule, MatCardModule, MatDividerModule],
  exports: [UserDetailCardComponent]
})
export class UserDetailCardModule {}
