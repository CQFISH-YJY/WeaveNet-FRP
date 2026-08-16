import{_ as J}from"./SvgIcon-pf2fZObV.js";import{G as xe,H as ke,L as Z,N as i,O as q,K as A,M as g,aH as Q,R as _e,b6 as K,J as l,S as B,T as Se,b7 as Ce,W as Be,Z as oe,Y as Re,l as j,$ as $e,v as M,a1 as ze,a2 as E,a3 as R,b8 as I,b9 as p,b as Ve,y as Me,o as m,d as V,e as c,g as Y,w as ee,j as te,h as T,B as Ne,F as ae,s as ne,ba as Fe,k as Pe,t as W,m as G,a5 as Te,bb as We}from"./index-BNsw-iSP.js";import{_ as je}from"./_plugin-vue_export-helper-DlAUqK2U.js";import{u as Oe}from"./use-merged-state-CeK_hdRR.js";import{N as He}from"./InputNumber-C6NMzErY.js";import{N as Ue}from"./Input-CMAj-uXk.js";const Le={buttonHeightSmall:"14px",buttonHeightMedium:"18px",buttonHeightLarge:"22px",buttonWidthSmall:"14px",buttonWidthMedium:"18px",buttonWidthLarge:"22px",buttonWidthPressedSmall:"20px",buttonWidthPressedMedium:"24px",buttonWidthPressedLarge:"28px",railHeightSmall:"18px",railHeightMedium:"22px",railHeightLarge:"26px",railWidthSmall:"32px",railWidthMedium:"40px",railWidthLarge:"48px"};function De(e){const{primaryColor:d,opacityDisabled:u,borderRadius:r,textColor3:v}=e;return Object.assign(Object.assign({},Le),{iconColor:v,textColor:"white",loadingColor:d,opacityDisabled:u,railColor:"rgba(0, 0, 0, .14)",railColorActive:d,buttonBoxShadow:"0 1px 4px 0 rgba(0, 0, 0, 0.3), inset 0 0 1px 0 rgba(0, 0, 0, 0.05)",buttonColor:"#FFF",railBorderRadiusSmall:r,railBorderRadiusMedium:r,railBorderRadiusLarge:r,buttonBorderRadiusSmall:r,buttonBorderRadiusMedium:r,buttonBorderRadiusLarge:r,boxShadowFocus:`0 0 0 2px ${ke(d,{alpha:.2})}`})}const Ae={common:xe,self:De},Ke=Z("switch",`
 height: var(--n-height);
 min-width: var(--n-width);
 vertical-align: middle;
 user-select: none;
 -webkit-user-select: none;
 display: inline-flex;
 outline: none;
 justify-content: center;
 align-items: center;
`,[i("children-placeholder",`
 height: var(--n-rail-height);
 display: flex;
 flex-direction: column;
 overflow: hidden;
 pointer-events: none;
 visibility: hidden;
 `),i("rail-placeholder",`
 display: flex;
 flex-wrap: none;
 `),i("button-placeholder",`
 width: calc(1.75 * var(--n-rail-height));
 height: var(--n-rail-height);
 `),Z("base-loading",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 font-size: calc(var(--n-button-width) - 4px);
 color: var(--n-loading-color);
 transition: color .3s var(--n-bezier);
 `,[q({left:"50%",top:"50%",originalTransform:"translateX(-50%) translateY(-50%)"})]),i("checked, unchecked",`
 transition: color .3s var(--n-bezier);
 color: var(--n-text-color);
 box-sizing: border-box;
 position: absolute;
 white-space: nowrap;
 top: 0;
 bottom: 0;
 display: flex;
 align-items: center;
 line-height: 1;
 `),i("checked",`
 right: 0;
 padding-right: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),i("unchecked",`
 left: 0;
 justify-content: flex-end;
 padding-left: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),A("&:focus",[i("rail",`
 box-shadow: var(--n-box-shadow-focus);
 `)]),g("round",[i("rail","border-radius: calc(var(--n-rail-height) / 2);",[i("button","border-radius: calc(var(--n-button-height) / 2);")])]),Q("disabled",[Q("icon",[g("rubber-band",[g("pressed",[i("rail",[i("button","max-width: var(--n-button-width-pressed);")])]),i("rail",[A("&:active",[i("button","max-width: var(--n-button-width-pressed);")])]),g("active",[g("pressed",[i("rail",[i("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])]),i("rail",[A("&:active",[i("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])])])])])]),g("active",[i("rail",[i("button","left: calc(100% - var(--n-button-width) - var(--n-offset))")])]),i("rail",`
 overflow: hidden;
 height: var(--n-rail-height);
 min-width: var(--n-rail-width);
 border-radius: var(--n-rail-border-radius);
 cursor: pointer;
 position: relative;
 transition:
 opacity .3s var(--n-bezier),
 background .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 background-color: var(--n-rail-color);
 `,[i("button-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 font-size: calc(var(--n-button-height) - 4px);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 display: flex;
 justify-content: center;
 align-items: center;
 line-height: 1;
 `,[q()]),i("button",`
 align-items: center; 
 top: var(--n-offset);
 left: var(--n-offset);
 height: var(--n-button-height);
 width: var(--n-button-width-pressed);
 max-width: var(--n-button-width);
 border-radius: var(--n-button-border-radius);
 background-color: var(--n-button-color);
 box-shadow: var(--n-button-box-shadow);
 box-sizing: border-box;
 cursor: inherit;
 content: "";
 position: absolute;
 transition:
 background-color .3s var(--n-bezier),
 left .3s var(--n-bezier),
 opacity .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 `)]),g("active",[i("rail","background-color: var(--n-rail-color-active);")]),g("loading",[i("rail",`
 cursor: wait;
 `)]),g("disabled",[i("rail",`
 cursor: not-allowed;
 opacity: .5;
 `)])]),Ee=Object.assign(Object.assign({},oe.props),{size:String,value:{type:[String,Number,Boolean],default:void 0},loading:Boolean,defaultValue:{type:[String,Number,Boolean],default:!1},disabled:{type:Boolean,default:void 0},round:{type:Boolean,default:!0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],checkedValue:{type:[String,Number,Boolean],default:!0},uncheckedValue:{type:[String,Number,Boolean],default:!1},railStyle:Function,rubberBand:{type:Boolean,default:!0},spinProps:Object,onChange:[Function,Array]});let F;const Ie=_e({name:"Switch",props:Ee,slots:Object,setup(e){F===void 0&&(typeof CSS<"u"?typeof CSS.supports<"u"?F=CSS.supports("width","max(1px)"):F=!1:F=!0);const{mergedClsPrefixRef:d,inlineThemeDisabled:u,mergedComponentPropsRef:r}=Be(e),v=oe("Switch","-switch",Ke,Ae,e,d),b=Re(e,{mergedSize(o){var w,x;if(e.size!==void 0)return e.size;if(o)return o.mergedSize.value;const C=(x=(w=r==null?void 0:r.value)===null||w===void 0?void 0:w.Switch)===null||x===void 0?void 0:x.size;return C||"medium"}}),{mergedSizeRef:S,mergedDisabledRef:f}=b,y=j(e.defaultValue),$=ze(e,"value"),a=Oe($,y),t=M(()=>a.value===e.checkedValue),n=j(!1),s=j(!1),h=M(()=>{const{railStyle:o}=e;if(o)return o({focused:s.value,checked:t.value})});function P(o){const{"onUpdate:value":w,onChange:x,onUpdateValue:C}=e,{nTriggerFormInput:O,nTriggerFormChange:H}=b;w&&E(w,o),C&&E(C,o),x&&E(x,o),y.value=o,O(),H()}function ie(){const{nTriggerFormFocus:o}=b;o()}function se(){const{nTriggerFormBlur:o}=b;o()}function re(){e.loading||f.value||(a.value!==e.checkedValue?P(e.checkedValue):P(e.uncheckedValue))}function le(){s.value=!0,ie()}function ce(){s.value=!1,se(),n.value=!1}function de(o){e.loading||f.value||o.key===" "&&(a.value!==e.checkedValue?P(e.checkedValue):P(e.uncheckedValue),n.value=!1)}function ue(o){e.loading||f.value||o.key===" "&&(o.preventDefault(),n.value=!0)}const X=M(()=>{const{value:o}=S,{self:{opacityDisabled:w,railColor:x,railColorActive:C,buttonBoxShadow:O,buttonColor:H,boxShadowFocus:he,loadingColor:be,textColor:fe,iconColor:ve,[R("buttonHeight",o)]:k,[R("buttonWidth",o)]:ge,[R("buttonWidthPressed",o)]:pe,[R("railHeight",o)]:_,[R("railWidth",o)]:N,[R("railBorderRadius",o)]:me,[R("buttonBorderRadius",o)]:ye},common:{cubicBezierEaseInOut:we}}=v.value;let U,L,D;return F?(U=`calc((${_} - ${k}) / 2)`,L=`max(${_}, ${k})`,D=`max(${N}, calc(${N} + ${k} - ${_}))`):(U=I((p(_)-p(k))/2),L=I(Math.max(p(_),p(k))),D=p(_)>p(k)?N:I(p(N)+p(k)-p(_))),{"--n-bezier":we,"--n-button-border-radius":ye,"--n-button-box-shadow":O,"--n-button-color":H,"--n-button-width":ge,"--n-button-width-pressed":pe,"--n-button-height":k,"--n-height":L,"--n-offset":U,"--n-opacity-disabled":w,"--n-rail-border-radius":me,"--n-rail-color":x,"--n-rail-color-active":C,"--n-rail-height":_,"--n-rail-width":N,"--n-width":D,"--n-box-shadow-focus":he,"--n-loading-color":be,"--n-text-color":fe,"--n-icon-color":ve}}),z=u?$e("switch",M(()=>S.value[0]),X,e):void 0;return{handleClick:re,handleBlur:ce,handleFocus:le,handleKeyup:de,handleKeydown:ue,mergedRailStyle:h,pressed:n,mergedClsPrefix:d,mergedValue:a,checked:t,mergedDisabled:f,cssVars:u?void 0:X,themeClass:z==null?void 0:z.themeClass,onRender:z==null?void 0:z.onRender}},render(){const{mergedClsPrefix:e,mergedDisabled:d,checked:u,mergedRailStyle:r,onRender:v,$slots:b}=this;v==null||v();const{checked:S,unchecked:f,icon:y,"checked-icon":$,"unchecked-icon":a}=b,t=!(K(y)&&K($)&&K(a));return l("div",{role:"switch","aria-checked":u,class:[`${e}-switch`,this.themeClass,t&&`${e}-switch--icon`,u&&`${e}-switch--active`,d&&`${e}-switch--disabled`,this.round&&`${e}-switch--round`,this.loading&&`${e}-switch--loading`,this.pressed&&`${e}-switch--pressed`,this.rubberBand&&`${e}-switch--rubber-band`],tabindex:this.mergedDisabled?void 0:0,style:this.cssVars,onClick:this.handleClick,onFocus:this.handleFocus,onBlur:this.handleBlur,onKeyup:this.handleKeyup,onKeydown:this.handleKeydown},l("div",{class:`${e}-switch__rail`,"aria-hidden":"true",style:r},B(S,n=>B(f,s=>n||s?l("div",{"aria-hidden":!0,class:`${e}-switch__children-placeholder`},l("div",{class:`${e}-switch__rail-placeholder`},l("div",{class:`${e}-switch__button-placeholder`}),n),l("div",{class:`${e}-switch__rail-placeholder`},l("div",{class:`${e}-switch__button-placeholder`}),s)):null)),l("div",{class:`${e}-switch__button`},B(y,n=>B($,s=>B(a,h=>l(Se,null,{default:()=>this.loading?l(Ce,Object.assign({key:"loading",clsPrefix:e,strokeWidth:20},this.spinProps)):this.checked&&(s||n)?l("div",{class:`${e}-switch__button-icon`,key:s?"checked-icon":"icon"},s||n):!this.checked&&(h||n)?l("div",{class:`${e}-switch__button-icon`,key:h?"unchecked-icon":"icon"},h||n):null})))),B(S,n=>n&&l("div",{key:"checked",class:`${e}-switch__checked`},n)),B(f,n=>n&&l("div",{key:"unchecked",class:`${e}-switch__unchecked`},n)))))}}),Ye={class:"page-container"},Ge={class:"page-title"},Xe={class:"group-head"},Je={class:"group-title"},Ze={class:"group-desc"},qe={class:"config-list"},Qe={class:"config-label"},et={class:"config-key"},tt={class:"config-input"},at={key:0,class:"empty-tip"},nt={__name:"Config",setup(e){const d=Ve(),u=j(!1),r=Pe({}),v={signin_points:"每日签到积分",continuous_days:"连续签到天数",continuous_bonus:"连续签到奖励",exchange_cost:"兑换所需积分",exchange_plan:"兑换套餐名称",smtp_host:"SMTP 服务器",smtp_port:"SMTP 端口",smtp_user:"SMTP 账号",smtp_pass:"SMTP 密码",smtp_from:"发件人地址",smtp_tls:"启用 TLS"},b=M(()=>Object.entries(r).map(([a,t])=>({key:a,value:t,type:typeof t=="number"?"number":typeof t=="boolean"?"boolean":"string"}))),S=M(()=>[{key:"signin",label:"签到规则",icon:"check",desc:"每日签到积分与连续签到奖励",match:t=>t.includes("signin")||t.includes("continuous")},{key:"exchange",label:"积分兑换",icon:"plan",desc:"积分兑换会员所需积分",match:t=>t.includes("exchange")},{key:"smtp",label:"邮件 SMTP",icon:"mail",desc:"发送验证码邮件的 SMTP 服务",match:t=>t.includes("smtp")},{key:"other",label:"其他配置",icon:"config",desc:"其余系统配置项",match:()=>!0}].map(t=>({...t,items:b.value.filter(n=>t.match(n.key))})).filter(t=>t.items.length));async function f(){try{const a=await Fe()||{};Object.keys(r).forEach(t=>delete r[t]),Array.isArray(a)?a.forEach(t=>{r[t.key]=y(t.value)}):Object.entries(a).forEach(([t,n])=>{r[t]=y(n)})}catch{}}function y(a){if(a==="true"||a==="false")return a==="true";const t=Number(a);return a!==""&&!Number.isNaN(t)&&typeof a=="string"&&a.trim()!==""?t:a}async function $(){u.value=!0;try{for(const a of b.value)await We({key:a.key,value:a.value});d.success("全部配置已保存"),await f()}catch{d.error("部分配置保存失败，请检查后重试")}finally{u.value=!1}}return Me(f),(a,t)=>(m(),V("div",Ye,[c("div",Ge,[t[1]||(t[1]=c("div",null,[c("h2",null,"系统配置"),c("div",{class:"sub"},"管理签到规则、积分兑换与邮件服务配置")],-1)),Y(T(Ne),{class:"btn-grad-cyan",loading:u.value,onClick:$},{icon:ee(()=>[Y(J,{name:"check",size:15})]),default:ee(()=>[t[0]||(t[0]=te(" 保存全部 ",-1))]),_:1},8,["loading"])]),(m(!0),V(ae,null,ne(S.value,n=>(m(),V("div",{key:n.key,class:"glass-card group-card"},[c("div",Xe,[c("span",Je,[Y(J,{name:n.icon,size:16,class:"group-icon"},null,8,["name"]),te(" "+W(n.label),1)]),c("span",Ze,W(n.desc),1)]),c("div",qe,[(m(!0),V(ae,null,ne(n.items,s=>(m(),V("div",{key:s.key,class:"config-row"},[c("div",Qe,[c("div",null,W(v[s.key]||s.key),1),c("div",et,W(s.key),1)]),c("div",tt,[s.type==="number"?(m(),G(T(He),{key:0,value:s.value,"onUpdate:value":h=>s.value=h,style:{width:"280px"},min:0},null,8,["value","onUpdate:value"])):s.type==="boolean"?(m(),G(T(Ie),{key:1,value:s.value,"onUpdate:value":h=>s.value=h},null,8,["value","onUpdate:value"])):(m(),G(T(Ue),{key:2,value:s.value,"onUpdate:value":h=>s.value=h,style:{width:"320px"}},null,8,["value","onUpdate:value"]))])]))),128)),n.items.length?Te("",!0):(m(),V("div",at,"暂无配置项"))])]))),128))]))}},dt=je(nt,[["__scopeId","data-v-370d3f74"]]);export{dt as default};
