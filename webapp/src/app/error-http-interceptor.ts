import { Injectable, Injector } from '@angular/core';
import { HttpEvent, HttpInterceptor, HttpHandler, HttpRequest } from '@angular/common/http';
import { Observable, of as observableOf, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { MatSnackBar } from '@angular/material';


@Injectable()
export class ErrorHttpInterceptor implements HttpInterceptor {
    constructor(
        private toast: MatSnackBar
    ) { }

    getMessage(res) {
        if (res.error && res.error.error) {
            return res.error.error;
        }

        switch (String(res.status)) {
            case '404':
                return 'Resource not found';
            case '400':
                return 'Bad request';
            default:
                return 'Internal server error';
        }
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        return next.handle(req).pipe(catchError((err) => {
            this.toast.open(this.getMessage(err), undefined, {
                horizontalPosition: 'right',
                duration: 5000,
                politeness: 'assertive',
                panelClass: 'snackbar-error',
            });
            return throwError(err);
        }));
    }
}
