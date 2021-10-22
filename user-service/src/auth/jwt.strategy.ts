import { HttpException, HttpStatus, Injectable } from '@nestjs/common';
import { PassportStrategy } from '@nestjs/passport';
import { ExtractJwt, Strategy } from 'passport-jwt';

import {configService} from '../config/config.service';
import { AuthorizedUserDto } from '../user/user.dto';
import { JwtPayload } from '../user/interfaces/jwt-payload.interface';
import { AuthService } from './auth.service';


@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
    constructor(private readonly authService: AuthService) {
        super({
            jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
            ignoreExpiration: false,
            secretOrKey: configService.getValue('SECRET_KEY')
        });
    }

    async validate(payload: JwtPayload): Promise<AuthorizedUserDto> {
        const user = await this.authService.validateUser(payload);
        console.log(user)
        if (!user) {
            throw new HttpException('Invalid token',
                HttpStatus.UNAUTHORIZED);
        }

        return user;
    }
}
