import { BadRequestException, Inject, Injectable } from '@nestjs/common';
import { Pool } from 'pg';
import { AccountRole } from '../auth/dto';
import { DATABASE_POOL } from '../database/database.constants';
import { UpdateProfileDto } from './profile.dto';
@Injectable()
export class ProfilesRepository{
  constructor(@Inject(DATABASE_POOL)private readonly pool:Pool){}
  async get(userId:string,role:AccountRole){const table=role===AccountRole.BUYER?'buyer_profiles':'contributor_profiles';const result=await this.pool.query(`SELECT * FROM ${table} WHERE user_id=$1`,[userId]);return result.rows[0]??{user_id:userId};}
  async update(userId:string,role:AccountRole,input:UpdateProfileDto){
    if(role===AccountRole.BUYER){
      if(input.profession!==undefined||input.biography!==undefined||input.city!==undefined)throw new BadRequestException('Contributor fields are not allowed for buyer profiles');
      const result=await this.pool.query(`INSERT INTO buyer_profiles(user_id,display_name,organisation,country_code) VALUES($1,$2,$3,$4) ON CONFLICT(user_id) DO UPDATE SET display_name=COALESCE(EXCLUDED.display_name,buyer_profiles.display_name),organisation=COALESCE(EXCLUDED.organisation,buyer_profiles.organisation),country_code=COALESCE(EXCLUDED.country_code,buyer_profiles.country_code),updated_at=now() RETURNING *`,[userId,input.displayName??null,input.organisation??null,input.countryCode??null]);return result.rows[0];
    }
    if(input.organisation!==undefined)throw new BadRequestException('Buyer fields are not allowed for contributor profiles');
    const result=await this.pool.query(`INSERT INTO contributor_profiles(user_id,display_name,profession,biography,country_code,city) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(user_id) DO UPDATE SET display_name=COALESCE(EXCLUDED.display_name,contributor_profiles.display_name),profession=COALESCE(EXCLUDED.profession,contributor_profiles.profession),biography=COALESCE(EXCLUDED.biography,contributor_profiles.biography),country_code=COALESCE(EXCLUDED.country_code,contributor_profiles.country_code),city=COALESCE(EXCLUDED.city,contributor_profiles.city),updated_at=now() RETURNING *`,[userId,input.displayName??null,input.profession??null,input.biography??null,input.countryCode??null,input.city??null]);return result.rows[0];
  }
}
