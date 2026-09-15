const permittedFields=new Set(['timestamp','severity','service','event','correlationId','method','route','status','durationMs','environment','count']);
const permittedSeverities=new Set(['debug','info','warn','error','fatal']);

export function writeStructuredLog(fields:Record<string,unknown>):void{
  const entry:Record<string,unknown>={timestamp:new Date().toISOString(),severity:'info',service:'tamo-auth-service'};
  for(const [key,value] of Object.entries(fields))if(permittedFields.has(key)&&value!==undefined)entry[key]=value;
  if(!permittedSeverities.has(String(entry.severity)))entry.severity='info';
  process.stdout.write(`${JSON.stringify(entry)}\n`);
}
