import { Component, OnInit } from '@angular/core';
import { User } from 'ecad-angular-auth';
import { Router } from '@angular/router';

@Component({
  selector: 'app-protected',
  templateUrl: './protected.component.html',
  styleUrls: ['./protected.component.scss']
})
export class ProtectedComponent implements OnInit {

  constructor(
    private router: Router
  ) { }

  ngOnInit() {
  }

  userClicked($event: User) {
    this.router.navigateByUrl(`/user/${$event.id}`);
  }

}
