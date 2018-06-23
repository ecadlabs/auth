export interface UserToken {
    exp: number;
    'http://localhost:8000/email': string;
    'http://localhost:8000/name': string;
    'http://localhost:8000/roles': any[];
    iat: number;
    iss: string;
    sub: string;
}
