import { Component, OnInit, Input } from '@angular/core';
import { User } from '../../ecad-angular-auth-admin/interfaces/user.i';

@Component({
  selector: 'auth-user-detail-card',
  templateUrl: './user-detail-card.component.html',
  styleUrls: ['./user-detail-card.component.scss']
})
export class UserDetailCardComponent implements OnInit {
  @Input()
  user: User;

  constructor() {}

  ngOnInit() {}
}
