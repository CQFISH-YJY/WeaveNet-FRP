import{_ as E}from"./SvgIcon-pf2fZObV.js";import{S as we}from"./StatusBadge-BNR7dHAk.js";import{G as ze,H as Te,I as Se,J as _,K as S,L as b,M as O,N as V,O as Re,P as $e,Q as De,R as Ne,S as Me,T as Ue,U as Pe,V as Be,W as Ie,l as R,X as Ve,Y as Fe,Z as ce,_ as He,$ as je,v as ae,a0 as Ke,a1 as Oe,a2 as ee,a3 as ne,b as Le,a4 as Ee,y as qe,o as H,d as G,e as n,g as r,w as i,j as g,h as t,B as M,F as We,s as Ae,t as x,m as J,a5 as oe,a6 as te,a7 as Ge,a8 as Je,a9 as Xe,aa as Ye,ab as Qe,ac as Ze,k as eo,ad as oo,ae as lo}from"./index-BNsw-iSP.js";import{t as ao,b as no,c as to,f as re}from"./format-BGKMqIAB.js";import{_ as ro}from"./_plugin-vue_export-helper-DlAUqK2U.js";import{N as so}from"./Table-MgfdnSBg.js";import{N as se}from"./Select-6vhbLWFw.js";import{N as ie}from"./InputNumber-C6NMzErY.js";import{u as io}from"./use-merged-state-CeK_hdRR.js";import{a as co,N as j}from"./FormItem-DRDr7xL5.js";import{N as X}from"./Input-CMAj-uXk.js";import"./Popover-rVfyQ8tN.js";import"./get-0HCPGJU3.js";const uo={sizeSmall:"14px",sizeMedium:"16px",sizeLarge:"18px",labelPadding:"0 8px",labelFontWeight:"400"};function bo(c){const{baseColor:d,inputColorDisabled:w,cardColor:y,modalColor:$,popoverColor:U,textColorDisabled:p,borderColor:h,primaryColor:m,textColor2:u,fontSizeSmall:v,fontSizeMedium:D,fontSizeLarge:N,borderRadiusSmall:a,lineHeight:P}=c;return Object.assign(Object.assign({},uo),{labelLineHeight:P,fontSizeSmall:v,fontSizeMedium:D,fontSizeLarge:N,borderRadius:a,color:d,colorChecked:m,colorDisabled:w,colorDisabledChecked:w,colorTableHeader:y,colorTableHeaderModal:$,colorTableHeaderPopover:U,checkMarkColor:d,checkMarkColorDisabled:p,checkMarkColorDisabledChecked:p,border:`1px solid ${h}`,borderDisabled:`1px solid ${h}`,borderDisabledChecked:`1px solid ${h}`,borderChecked:`1px solid ${m}`,borderFocus:`1px solid ${m}`,boxShadowFocus:`0 0 0 2px ${Te(m,{alpha:.3})}`,textColor:u,textColorDisabled:p})}const po={common:ze,self:bo},mo=Se("n-checkbox-group"),ho=()=>_("svg",{viewBox:"0 0 64 64",class:"check-icon"},_("path",{d:"M50.42,16.76L22.34,39.45l-8.1-11.46c-1.12-1.58-3.3-1.96-4.88-0.84c-1.58,1.12-1.95,3.3-0.84,4.88l10.26,14.51  c0.56,0.79,1.42,1.31,2.38,1.45c0.16,0.02,0.32,0.03,0.48,0.03c0.8,0,1.57-0.27,2.2-0.78l30.99-25.03c1.5-1.21,1.74-3.42,0.52-4.92  C54.13,15.78,51.93,15.55,50.42,16.76z"})),vo=()=>_("svg",{viewBox:"0 0 100 100",class:"line-icon"},_("path",{d:"M80.2,55.5H21.4c-2.8,0-5.1-2.5-5.1-5.5l0,0c0-3,2.3-5.5,5.1-5.5h58.7c2.8,0,5.1,2.5,5.1,5.5l0,0C85.2,53.1,82.9,55.5,80.2,55.5z"})),fo=S([b("checkbox",`
 font-size: var(--n-font-size);
 outline: none;
 cursor: pointer;
 display: inline-flex;
 flex-wrap: nowrap;
 align-items: flex-start;
 word-break: break-word;
 line-height: var(--n-size);
 --n-merged-color-table: var(--n-color-table);
 `,[O("show-label","line-height: var(--n-label-line-height);"),S("&:hover",[b("checkbox-box",[V("border","border: var(--n-border-checked);")])]),S("&:focus:not(:active)",[b("checkbox-box",[V("border",`
 border: var(--n-border-focus);
 box-shadow: var(--n-box-shadow-focus);
 `)])]),O("inside-table",[b("checkbox-box",`
 background-color: var(--n-merged-color-table);
 `)]),O("checked",[b("checkbox-box",`
 background-color: var(--n-color-checked);
 `,[b("checkbox-icon",[S(".check-icon",`
 opacity: 1;
 transform: scale(1);
 `)])])]),O("indeterminate",[b("checkbox-box",[b("checkbox-icon",[S(".check-icon",`
 opacity: 0;
 transform: scale(.5);
 `),S(".line-icon",`
 opacity: 1;
 transform: scale(1);
 `)])])]),O("checked, indeterminate",[S("&:focus:not(:active)",[b("checkbox-box",[V("border",`
 border: var(--n-border-checked);
 box-shadow: var(--n-box-shadow-focus);
 `)])]),b("checkbox-box",`
 background-color: var(--n-color-checked);
 border-left: 0;
 border-top: 0;
 `,[V("border",{border:"var(--n-border-checked)"})])]),O("disabled",{cursor:"not-allowed"},[O("checked",[b("checkbox-box",`
 background-color: var(--n-color-disabled-checked);
 `,[V("border",{border:"var(--n-border-disabled-checked)"}),b("checkbox-icon",[S(".check-icon, .line-icon",{fill:"var(--n-check-mark-color-disabled-checked)"})])])]),b("checkbox-box",`
 background-color: var(--n-color-disabled);
 `,[V("border",`
 border: var(--n-border-disabled);
 `),b("checkbox-icon",[S(".check-icon, .line-icon",`
 fill: var(--n-check-mark-color-disabled);
 `)])]),V("label",`
 color: var(--n-text-color-disabled);
 `)]),b("checkbox-box-wrapper",`
 position: relative;
 width: var(--n-size);
 flex-shrink: 0;
 flex-grow: 0;
 user-select: none;
 -webkit-user-select: none;
 `),b("checkbox-box",`
 position: absolute;
 left: 0;
 top: 50%;
 transform: translateY(-50%);
 height: var(--n-size);
 width: var(--n-size);
 display: inline-block;
 box-sizing: border-box;
 border-radius: var(--n-border-radius);
 background-color: var(--n-color);
 transition: background-color 0.3s var(--n-bezier);
 `,[V("border",`
 transition:
 border-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 border-radius: inherit;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 border: var(--n-border);
 `),b("checkbox-icon",`
 display: flex;
 align-items: center;
 justify-content: center;
 position: absolute;
 left: 1px;
 right: 1px;
 top: 1px;
 bottom: 1px;
 `,[S(".check-icon, .line-icon",`
 width: 100%;
 fill: var(--n-check-mark-color);
 opacity: 0;
 transform: scale(0.5);
 transform-origin: center;
 transition:
 fill 0.3s var(--n-bezier),
 transform 0.3s var(--n-bezier),
 opacity 0.3s var(--n-bezier),
 border-color 0.3s var(--n-bezier);
 `),Re({left:"1px",top:"1px"})])]),V("label",`
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 `,[S("&:empty",{display:"none"})])]),$e(b("checkbox",`
 --n-merged-color-table: var(--n-color-table-modal);
 `)),De(b("checkbox",`
 --n-merged-color-table: var(--n-color-table-popover);
 `))]),ko=Object.assign(Object.assign({},ce.props),{size:String,checked:{type:[Boolean,String,Number],default:void 0},defaultChecked:{type:[Boolean,String,Number],default:!1},value:[String,Number],disabled:{type:Boolean,default:void 0},indeterminate:Boolean,label:String,focusable:{type:Boolean,default:!0},checkedValue:{type:[Boolean,String,Number],default:!0},uncheckedValue:{type:[Boolean,String,Number],default:!1},"onUpdate:checked":[Function,Array],onUpdateChecked:[Function,Array],privateInsideTable:Boolean,onChange:[Function,Array]}),le=Ne({name:"Checkbox",props:ko,setup(c){const d=Be(mo,null),w=R(null),{mergedClsPrefixRef:y,inlineThemeDisabled:$,mergedRtlRef:U,mergedComponentPropsRef:p}=Ie(c),h=R(c.defaultChecked),m=Oe(c,"checked"),u=io(m,h),v=Ve(()=>{if(d){const s=d.valueSetRef.value;return s&&c.value!==void 0?s.has(c.value):!1}else return u.value===c.checkedValue}),D=Fe(c,{mergedSize(s){var f,k;const{size:C}=c;if(C!==void 0)return C;if(d){const{value:T}=d.mergedSizeRef;if(T!==void 0)return T}if(s){const{mergedSize:T}=s;if(T!==void 0)return T.value}const z=(k=(f=p==null?void 0:p.value)===null||f===void 0?void 0:f.Checkbox)===null||k===void 0?void 0:k.size;return z||"medium"},mergedDisabled(s){const{disabled:f}=c;if(f!==void 0)return f;if(d){if(d.disabledRef.value)return!0;const{maxRef:{value:k},checkedCountRef:C}=d;if(k!==void 0&&C.value>=k&&!v.value)return!0;const{minRef:{value:z}}=d;if(z!==void 0&&C.value<=z&&v.value)return!0}return s?s.disabled.value:!1}}),{mergedDisabledRef:N,mergedSizeRef:a}=D,P=ce("Checkbox","-checkbox",fo,po,c,y);function B(s){if(d&&c.value!==void 0)d.toggleCheckbox(!v.value,c.value);else{const{onChange:f,"onUpdate:checked":k,onUpdateChecked:C}=c,{nTriggerFormInput:z,nTriggerFormChange:T}=D,F=v.value?c.uncheckedValue:c.checkedValue;k&&ee(k,F,s),C&&ee(C,F,s),f&&ee(f,F,s),z(),T(),h.value=F}}function Y(s){N.value||B(s)}function q(s){if(!N.value)switch(s.key){case" ":case"Enter":B(s)}}function Q(s){switch(s.key){case" ":s.preventDefault()}}const Z={focus:()=>{var s;(s=w.value)===null||s===void 0||s.focus()},blur:()=>{var s;(s=w.value)===null||s===void 0||s.blur()}},K=He("Checkbox",U,y),W=ae(()=>{const{value:s}=a,{common:{cubicBezierEaseInOut:f},self:{borderRadius:k,color:C,colorChecked:z,colorDisabled:T,colorTableHeader:F,colorTableHeaderModal:o,colorTableHeaderPopover:e,checkMarkColor:l,checkMarkColorDisabled:A,border:L,borderFocus:de,borderDisabled:ue,borderChecked:be,boxShadowFocus:pe,textColor:me,textColorDisabled:he,checkMarkColorDisabledChecked:ve,colorDisabledChecked:fe,borderDisabledChecked:ke,labelPadding:ge,labelLineHeight:xe,labelFontWeight:ye,[ne("fontSize",s)]:Ce,[ne("size",s)]:_e}}=P.value;return{"--n-label-line-height":xe,"--n-label-font-weight":ye,"--n-size":_e,"--n-bezier":f,"--n-border-radius":k,"--n-border":L,"--n-border-checked":be,"--n-border-focus":de,"--n-border-disabled":ue,"--n-border-disabled-checked":ke,"--n-box-shadow-focus":pe,"--n-color":C,"--n-color-checked":z,"--n-color-table":F,"--n-color-table-modal":o,"--n-color-table-popover":e,"--n-color-disabled":T,"--n-color-disabled-checked":fe,"--n-text-color":me,"--n-text-color-disabled":he,"--n-check-mark-color":l,"--n-check-mark-color-disabled":A,"--n-check-mark-color-disabled-checked":ve,"--n-font-size":Ce,"--n-label-padding":ge}}),I=$?je("checkbox",ae(()=>a.value[0]),W,c):void 0;return Object.assign(D,Z,{rtlEnabled:K,selfRef:w,mergedClsPrefix:y,mergedDisabled:N,renderedChecked:v,mergedTheme:P,labelId:Ke(),handleClick:Y,handleKeyUp:q,handleKeyDown:Q,cssVars:$?void 0:W,themeClass:I==null?void 0:I.themeClass,onRender:I==null?void 0:I.onRender})},render(){var c;const{$slots:d,renderedChecked:w,mergedDisabled:y,indeterminate:$,privateInsideTable:U,cssVars:p,labelId:h,label:m,mergedClsPrefix:u,focusable:v,handleKeyUp:D,handleKeyDown:N,handleClick:a}=this;(c=this.onRender)===null||c===void 0||c.call(this);const P=Me(d.default,B=>m||B?_("span",{class:`${u}-checkbox__label`,id:h},m||B):null);return _("div",{ref:"selfRef",class:[`${u}-checkbox`,this.themeClass,this.rtlEnabled&&`${u}-checkbox--rtl`,w&&`${u}-checkbox--checked`,y&&`${u}-checkbox--disabled`,$&&`${u}-checkbox--indeterminate`,U&&`${u}-checkbox--inside-table`,P&&`${u}-checkbox--show-label`],tabindex:y||!v?void 0:0,role:"checkbox","aria-checked":$?"mixed":w,"aria-labelledby":h,style:p,onKeyup:D,onKeydown:N,onClick:a,onMousedown:()=>{Pe("selectstart",window,B=>{B.preventDefault()},{once:!0})}},_("div",{class:`${u}-checkbox-box-wrapper`}," ",_("div",{class:`${u}-checkbox-box`},_(Ue,null,{default:()=>this.indeterminate?_("div",{key:"indeterminate",class:`${u}-checkbox-icon`},vo()):_("div",{key:"check",class:`${u}-checkbox-icon`},ho())}),_("div",{class:`${u}-checkbox-box__border`}))),P)}}),go={class:"page-container"},xo={class:"page-title"},yo={class:"glass-card table-wrap"},Co={class:"cell-main"},_o={class:"tunnel-name"},wo={class:"cell-sub"},zo={class:"type-tag"},To={class:"cell-main"},So={class:"public-url"},Ro={class:"cell-sub"},$o={class:"cell-sub"},Do={class:"cell-sub"},No={class:"table-actions"},Mo={key:0},Uo={colspan:"9"},Po={class:"empty-tip"},Bo={class:"form-grid"},Io={class:"switch-row"},Vo={class:"modal-actions"},Fo={class:"code-block"},Ho={class:"modal-actions"},jo={__name:"Tunnels",setup(c){const d=Le(),w=Ee(),y=R([]),$=R([]),U=R(!1),p=R(!1),h=R(!1),m=R(null),u=R(!1),v=R(""),D=R(null),N=()=>({name:"",type:"tcp",node_id:null,local_ip:"127.0.0.1",local_port:null,remote_port:null,subdomain:"",custom_domain:"",kcp:!1,encryption:!1,compression:!1}),a=eo(N()),P=[{label:"TCP",value:"tcp"},{label:"UDP",value:"udp"},{label:"HTTP",value:"http"},{label:"HTTPS",value:"https"},{label:"STCP 安全隧道",value:"stcp"},{label:"XTCP 点对点",value:"xtcp"},{label:"KCP 加速",value:"kcp"},{label:"负载均衡",value:"loadbalance"}],B=ae(()=>$.value.map(o=>({label:`${o.name}${o.status==="online"||o.status===1?"":"（离线）"}`,value:o.id}))),Y={name:[{required:!0,message:"请输入隧道名称",trigger:["input","blur"]}],type:[{required:!0,message:"请选择隧道类型",trigger:["change"]}],node_id:[{required:!0,message:"请选择穿透节点",trigger:["change"]}],local_ip:[{required:!0,message:"请输入本地 IP",trigger:["input","blur"]}],local_port:[{required:!0,message:"请输入本地端口",trigger:["change"]}]};function q(o){return o==="http"||o==="https"}function Q(o){const e=o.online??o.status;return e==="running"||e==="online"||e===!0||e===1||e==="1"}function Z(o){return o.public_addr?o.public_addr:o.public_url?o.public_url:o.full_domain?o.full_domain:o.subdomain?o.subdomain:o.custom_domain?o.custom_domain:"-"}async function K(){try{y.value=await Ge()||[]}catch{}}async function W(){U.value=!0;try{$.value=await Je()||[]}catch{}finally{U.value=!1}}function I(){m.value=null,Object.assign(a,N()),p.value=!0}function s(o){m.value=o,Object.assign(a,{name:o.name||"",type:o.type||"tcp",node_id:o.node_id??null,local_ip:o.local_ip||"127.0.0.1",local_port:o.local_port??null,remote_port:o.remote_port??null,subdomain:o.subdomain||"",custom_domain:o.custom_domain||"",kcp:!!o.kcp,encryption:!!o.encryption,compression:!!o.compression}),p.value=!0}async function f(){var e;try{await((e=D.value)==null?void 0:e.validate())}catch{return}u.value=!0;const o={name:a.name,type:a.type,node_id:a.node_id,local_ip:a.local_ip,local_port:a.local_port};a.remote_port&&(o.remote_port=a.remote_port),a.subdomain&&(o.subdomain=a.subdomain),a.custom_domain&&(o.custom_domain=a.custom_domain),a.kcp&&(o.kcp=!0),a.encryption&&(o.encryption=!0),a.compression&&(o.compression=!0);try{m.value?(await oo(m.value.id,o),d.success("隧道已更新")):(await lo(o),d.success("隧道创建成功")),p.value=!1,await K()}catch{}finally{u.value=!1}}async function k(o){try{await Xe(o.id),d.success("隧道已启动"),await K()}catch{}}async function C(o){try{await Ye(o.id),d.success("隧道已停止"),await K()}catch{}}function z(o){w.warning({title:"删除隧道",content:`确定要删除隧道「${o.name}」吗？删除后不可恢复。`,positiveText:"确认删除",negativeText:"取消",onPositiveClick:async()=>{try{await Ze(o.id),d.success("隧道已删除"),await K()}catch{}}})}async function T(o){try{const e=await Qe(o.id);v.value=typeof e=="string"?e:(e==null?void 0:e.config)||JSON.stringify(e,null,2),h.value=!0}catch{}}async function F(){try{await navigator.clipboard.writeText(v.value),d.success("配置已复制到剪贴板")}catch{d.error("复制失败，请手动选择复制")}}return qe(()=>{K(),W()}),(o,e)=>(H(),G("div",go,[n("div",xo,[e[15]||(e[15]=n("div",null,[n("h2",null,"隧道管理"),n("div",{class:"sub"},"创建并管理你的内网穿透隧道")],-1)),r(t(M),{class:"btn-grad-orange",onClick:I},{icon:i(()=>[r(E,{name:"plus",size:16})]),default:i(()=>[e[14]||(e[14]=g(" 创建隧道 ",-1))]),_:1})]),n("div",yo,[r(t(so),{bordered:!1,"single-line":!1,size:"small"},{default:i(()=>[e[23]||(e[23]=n("thead",null,[n("tr",null,[n("th",null,"名称"),n("th",null,"类型"),n("th",null,"节点"),n("th",null,"本地地址"),n("th",null,"远程端口"),n("th",null,"公网地址"),n("th",null,"状态"),n("th",null,"流量"),n("th",{style:{width:"260px"}},"操作")])],-1)),n("tbody",null,[(H(!0),G(We,null,Ae(y.value,l=>{var A;return H(),G("tr",{key:l.id},[n("td",Co,[n("div",_o,x(l.name),1),n("div",wo,x(l.id)+" · "+x(l.local_ip||"-"),1)]),n("td",null,[n("span",zo,x(t(ao)(l.type)),1)]),n("td",null,x(l.node_name||((A=l.node)==null?void 0:A.name)||"-"),1),n("td",null,x(l.local_ip||"-")+":"+x(l.local_port||"-"),1),n("td",null,x(l.remote_port||"自动"),1),n("td",To,[n("div",So,x(Z(l)),1),n("div",Ro,x(l.subdomain||l.custom_domain||""),1)]),n("td",null,[r(we,{type:t(to)(l.online??l.status),text:t(no)(l.online??l.status)},null,8,["type","text"])]),n("td",null,[n("div",$o,"入 "+x(t(re)(l.traffic_in)),1),n("div",Do,"出 "+x(t(re)(l.traffic_out)),1)]),n("td",null,[n("div",No,[Q(l)?(H(),J(t(M),{key:1,size:"tiny",type:"warning",ghost:"",onClick:L=>C(l)},{icon:i(()=>[r(E,{name:"stop",size:13})]),default:i(()=>[e[17]||(e[17]=g(" 停止 ",-1))]),_:1},8,["onClick"])):(H(),J(t(M),{key:0,size:"tiny",type:"success",ghost:"",onClick:L=>k(l)},{icon:i(()=>[r(E,{name:"play",size:13})]),default:i(()=>[e[16]||(e[16]=g(" 启动 ",-1))]),_:1},8,["onClick"])),r(t(M),{size:"tiny",onClick:L=>T(l)},{icon:i(()=>[r(E,{name:"copy",size:13})]),default:i(()=>[e[18]||(e[18]=g(" 配置 ",-1))]),_:1},8,["onClick"]),r(t(M),{size:"tiny",onClick:L=>s(l)},{icon:i(()=>[r(E,{name:"edit",size:13})]),default:i(()=>[e[19]||(e[19]=g(" 编辑 ",-1))]),_:1},8,["onClick"]),r(t(M),{size:"tiny",type:"error",ghost:"",onClick:L=>z(l)},{icon:i(()=>[r(E,{name:"trash",size:13})]),default:i(()=>[e[20]||(e[20]=g(" 删除 ",-1))]),_:1},8,["onClick"])])])])}),128)),y.value.length?oe("",!0):(H(),G("tr",Mo,[n("td",Uo,[n("div",Po,[e[22]||(e[22]=n("div",{style:{"margin-bottom":"10px"}},"还没有创建任何隧道",-1)),r(t(M),{class:"btn-grad-cyan",size:"small",onClick:I},{default:i(()=>[...e[21]||(e[21]=[g("立即创建第一条隧道",-1)])]),_:1})])])]))])]),_:1})]),r(t(te),{show:p.value,"onUpdate:show":e[12]||(e[12]=l=>p.value=l),"mask-closable":!1,preset:"card",title:m.value?"编辑隧道":"创建隧道",style:{width:"560px","max-width":"94vw"}},{default:i(()=>[r(t(co),{ref_key:"formRef",ref:D,model:a,rules:Y,"label-placement":"top",size:"medium"},{default:i(()=>[n("div",Bo,[r(t(j),{label:"隧道名称",path:"name",style:{"grid-column":"span 2"}},{default:i(()=>[r(t(X),{value:a.name,"onUpdate:value":e[0]||(e[0]=l=>a.name=l),placeholder:"例如：我的网站",clearable:""},null,8,["value"])]),_:1}),r(t(j),{label:"隧道类型",path:"type"},{default:i(()=>[r(t(se),{value:a.type,"onUpdate:value":e[1]||(e[1]=l=>a.type=l),options:P},null,8,["value"])]),_:1}),r(t(j),{label:"穿透节点",path:"node_id"},{default:i(()=>[r(t(se),{value:a.node_id,"onUpdate:value":e[2]||(e[2]=l=>a.node_id=l),options:B.value,loading:U.value,placeholder:"选择节点"},null,8,["value","options","loading"])]),_:1}),r(t(j),{label:"本地 IP",path:"local_ip"},{default:i(()=>[r(t(X),{value:a.local_ip,"onUpdate:value":e[3]||(e[3]=l=>a.local_ip=l),placeholder:"127.0.0.1"},null,8,["value"])]),_:1}),r(t(j),{label:"本地端口",path:"local_port"},{default:i(()=>[r(t(ie),{value:a.local_port,"onUpdate:value":e[4]||(e[4]=l=>a.local_port=l),min:1,max:65535,style:{width:"100%"},placeholder:"例如 8080"},null,8,["value"])]),_:1}),r(t(j),{label:"远程端口",path:"remote_port"},{default:i(()=>[r(t(ie),{value:a.remote_port,"onUpdate:value":e[5]||(e[5]=l=>a.remote_port=l),min:1,max:65535,style:{width:"100%"},placeholder:"留空自动分配"},null,8,["value"])]),_:1}),q(a.type)?(H(),J(t(j),{key:0,label:"子域名（http/https）",path:"subdomain"},{default:i(()=>[r(t(X),{value:a.subdomain,"onUpdate:value":e[6]||(e[6]=l=>a.subdomain=l),placeholder:"例如 myweb"},null,8,["value"])]),_:1})):oe("",!0),q(a.type)?(H(),J(t(j),{key:1,label:"自定义域名（可选）",path:"custom_domain"},{default:i(()=>[r(t(X),{value:a.custom_domain,"onUpdate:value":e[7]||(e[7]=l=>a.custom_domain=l),placeholder:"例如 www.example.com"},null,8,["value"])]),_:1})):oe("",!0)]),n("div",Io,[r(t(le),{checked:a.kcp,"onUpdate:checked":e[8]||(e[8]=l=>a.kcp=l)},{default:i(()=>[...e[24]||(e[24]=[g("启用 KCP 加速",-1)])]),_:1},8,["checked"]),r(t(le),{checked:a.encryption,"onUpdate:checked":e[9]||(e[9]=l=>a.encryption=l)},{default:i(()=>[...e[25]||(e[25]=[g("加密传输",-1)])]),_:1},8,["checked"]),r(t(le),{checked:a.compression,"onUpdate:checked":e[10]||(e[10]=l=>a.compression=l)},{default:i(()=>[...e[26]||(e[26]=[g("压缩传输",-1)])]),_:1},8,["checked"])]),n("div",Vo,[r(t(M),{onClick:e[11]||(e[11]=l=>p.value=!1)},{default:i(()=>[...e[27]||(e[27]=[g("取消",-1)])]),_:1}),r(t(M),{class:"btn-grad-cyan",loading:u.value,onClick:f},{default:i(()=>[...e[28]||(e[28]=[g("保存",-1)])]),_:1},8,["loading"])])]),_:1},8,["model"])]),_:1},8,["show","title"]),r(t(te),{show:h.value,"onUpdate:show":e[13]||(e[13]=l=>h.value=l),preset:"card",title:"隧道配置",style:{width:"620px","max-width":"94vw"}},{default:i(()=>[n("div",Fo,x(v.value),1),n("div",Ho,[r(t(M),{class:"btn-grad-cyan",onClick:F},{default:i(()=>[...e[29]||(e[29]=[g("复制配置",-1)])]),_:1})])]),_:1},8,["show"])]))}},el=ro(jo,[["__scopeId","data-v-dada8d88"]]);export{el as default};
