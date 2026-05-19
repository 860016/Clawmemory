import{a as s}from"./go-client-gVAZVHPI.js";const e={getOverview:()=>s.get("/stats"),getUsage:t=>s.get("/stats/usage",{params:{days:t}})};export{e as s};
