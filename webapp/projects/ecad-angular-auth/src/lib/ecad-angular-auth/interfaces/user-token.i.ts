export interface UserToken {
    exp: number;
    email: string;
    name: string;
    roles: any[];
    iat: number;
    iss: string;
    sub: string;
}
