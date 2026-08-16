import{p as ze,B as Ke,V as _e,a as De,h as ce,r as Oe,N as $e,b as fe,c as Be}from"./Popover-B8QIQbLq.js";import{ar as Ae,as as Fe,af as Te,at as V,au as He,k as Le,U as G,av as de,l as T,R as _,J as s,aw as Me,G as he,H as je,I as ae,ax as X,V as H,L as k,M as N,K as $,ay as ve,az as le,W as me,Z,$ as be,v as w,aA as We,aB as we,X as q,aC as j,F as Ee,aD as Ue,aE as Ve,aF as Ge,aG as qe,aH as ue,N as O,aI as Xe,a1 as K,a2 as re,a3 as F}from"./index-CF_lMkBz.js";import{f as Ze}from"./get-D6zmsx4k.js";import{u as Je}from"./use-merged-state-Zkfs3Ins.js";import{c as Qe}from"./create-ref-setter-C4J8sofl.js";function Ye(e={},n){const i=Le({ctrl:!1,command:!1,win:!1,shift:!1,tab:!1}),{keydown:r,keyup:t}=e,o=a=>{switch(a.key){case"Control":i.ctrl=!0;break;case"Meta":i.command=!0,i.win=!0;break;case"Shift":i.shift=!0;break;case"Tab":i.tab=!0;break}r!==void 0&&Object.keys(r).forEach(b=>{if(b!==a.key)return;const v=r[b];if(typeof v=="function")v(a);else{const{stop:g=!1,prevent:x=!1}=v;g&&a.stopPropagation(),x&&a.preventDefault(),v.handler(a)}})},l=a=>{switch(a.key){case"Control":i.ctrl=!1;break;case"Meta":i.command=!1,i.win=!1;break;case"Shift":i.shift=!1;break;case"Tab":i.tab=!1;break}t!==void 0&&Object.keys(t).forEach(b=>{if(b!==a.key)return;const v=t[b];if(typeof v=="function")v(a);else{const{stop:g=!1,prevent:x=!1}=v;g&&a.stopPropagation(),x&&a.preventDefault(),v.handler(a)}})},u=()=>{(n===void 0||n.value)&&(G("keydown",document,o),G("keyup",document,l)),n!==void 0&&de(n,a=>{a?(G("keydown",document,o),G("keyup",document,l)):(V("keydown",document,o),V("keyup",document,l))})};return Ae()?(Fe(u),Te(()=>{(n===void 0||n.value)&&(V("keydown",document,o),V("keyup",document,l))})):u(),He(i)}function eo(e,n,i){const r=T(e.value);let t=null;return de(e,o=>{t!==null&&window.clearTimeout(t),o===!0?i&&!i.value?r.value=!0:t=window.setTimeout(()=>{r.value=!0},n):r.value=!1}),r}const oo=_({name:"ChevronRight",render(){return s("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},s("path",{d:"M5.64645 3.14645C5.45118 3.34171 5.45118 3.65829 5.64645 3.85355L9.79289 8L5.64645 12.1464C5.45118 12.3417 5.45118 12.6583 5.64645 12.8536C5.84171 13.0488 6.15829 13.0488 6.35355 12.8536L10.8536 8.35355C11.0488 8.15829 11.0488 7.84171 10.8536 7.64645L6.35355 3.14645C6.15829 2.95118 5.84171 2.95118 5.64645 3.14645Z",fill:"currentColor"}))}}),no={padding:"4px 0",optionIconSizeSmall:"14px",optionIconSizeMedium:"16px",optionIconSizeLarge:"16px",optionIconSizeHuge:"18px",optionSuffixWidthSmall:"14px",optionSuffixWidthMedium:"14px",optionSuffixWidthLarge:"16px",optionSuffixWidthHuge:"16px",optionIconSuffixWidthSmall:"32px",optionIconSuffixWidthMedium:"32px",optionIconSuffixWidthLarge:"36px",optionIconSuffixWidthHuge:"36px",optionPrefixWidthSmall:"14px",optionPrefixWidthMedium:"14px",optionPrefixWidthLarge:"16px",optionPrefixWidthHuge:"16px",optionIconPrefixWidthSmall:"36px",optionIconPrefixWidthMedium:"36px",optionIconPrefixWidthLarge:"40px",optionIconPrefixWidthHuge:"40px"};function to(e){const{primaryColor:n,textColor2:i,dividerColor:r,hoverColor:t,popoverColor:o,invertedColor:l,borderRadius:u,fontSizeSmall:a,fontSizeMedium:b,fontSizeLarge:v,fontSizeHuge:g,heightSmall:x,heightMedium:C,heightLarge:P,heightHuge:I,textColor3:S,opacityDisabled:R}=e;return Object.assign(Object.assign({},no),{optionHeightSmall:x,optionHeightMedium:C,optionHeightLarge:P,optionHeightHuge:I,borderRadius:u,fontSizeSmall:a,fontSizeMedium:b,fontSizeLarge:v,fontSizeHuge:g,optionTextColor:i,optionTextColorHover:i,optionTextColorActive:n,optionTextColorChildActive:n,color:o,dividerColor:r,suffixColor:i,prefixColor:i,optionColorHover:t,optionColorActive:je(n,{alpha:.1}),groupHeaderTextColor:S,optionTextColorInverted:"#BBB",optionTextColorHoverInverted:"#FFF",optionTextColorActiveInverted:"#FFF",optionTextColorChildActiveInverted:"#FFF",colorInverted:l,dividerColorInverted:"#BBB",suffixColorInverted:"#BBB",prefixColorInverted:"#BBB",optionColorHoverInverted:n,optionColorActiveInverted:n,groupHeaderTextColorInverted:"#AAA",optionOpacityDisabled:R})}const ro=Me({name:"Dropdown",common:he,peers:{Popover:ze},self:to}),se=ae("n-dropdown-menu"),J=ae("n-dropdown"),pe=ae("n-dropdown-option"),ye=_({name:"DropdownDivider",props:{clsPrefix:{type:String,required:!0}},render(){return s("div",{class:`${this.clsPrefix}-dropdown-divider`})}}),io=_({name:"DropdownGroupHeader",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0}},setup(){const{showIconRef:e,hasSubmenuRef:n}=H(se),{renderLabelRef:i,labelFieldRef:r,nodePropsRef:t,renderOptionRef:o}=H(J);return{labelField:r,showIcon:e,hasSubmenu:n,renderLabel:i,nodeProps:t,renderOption:o}},render(){var e;const{clsPrefix:n,hasSubmenu:i,showIcon:r,nodeProps:t,renderLabel:o,renderOption:l}=this,{rawNode:u}=this.tmNode,a=s("div",Object.assign({class:`${n}-dropdown-option`},t==null?void 0:t(u)),s("div",{class:`${n}-dropdown-option-body ${n}-dropdown-option-body--group`},s("div",{"data-dropdown-option":!0,class:[`${n}-dropdown-option-body__prefix`,r&&`${n}-dropdown-option-body__prefix--show-icon`]},X(u.icon)),s("div",{class:`${n}-dropdown-option-body__label`,"data-dropdown-option":!0},o?o(u):X((e=u.title)!==null&&e!==void 0?e:u[this.labelField])),s("div",{class:[`${n}-dropdown-option-body__suffix`,i&&`${n}-dropdown-option-body__suffix--has-submenu`],"data-dropdown-option":!0})));return l?l({node:a,option:u}):a}});function ao(e){const{textColorBase:n,opacity1:i,opacity2:r,opacity3:t,opacity4:o,opacity5:l}=e;return{color:n,opacity1Depth:i,opacity2Depth:r,opacity3Depth:t,opacity4Depth:o,opacity5Depth:l}}const lo={common:he,self:ao},so=k("icon",`
 height: 1em;
 width: 1em;
 line-height: 1em;
 text-align: center;
 display: inline-block;
 position: relative;
 fill: currentColor;
`,[N("color-transition",{transition:"color .3s var(--n-bezier)"}),N("depth",{color:"var(--n-color)"},[$("svg",{opacity:"var(--n-opacity)",transition:"opacity .3s var(--n-bezier)"})]),$("svg",{height:"1em",width:"1em"})]),co=Object.assign(Object.assign({},Z.props),{depth:[String,Number],size:[Number,String],color:String,component:[Object,Function]}),uo=_({_n_icon__:!0,name:"Icon",inheritAttrs:!1,props:co,setup(e){const{mergedClsPrefixRef:n,inlineThemeDisabled:i}=me(e),r=Z("Icon","-icon",so,lo,e,n),t=w(()=>{const{depth:l}=e,{common:{cubicBezierEaseInOut:u},self:a}=r.value;if(l!==void 0){const{color:b,[`opacity${l}Depth`]:v}=a;return{"--n-bezier":u,"--n-color":b,"--n-opacity":v}}return{"--n-bezier":u,"--n-color":"","--n-opacity":""}}),o=i?be("icon",w(()=>`${e.depth||"d"}`),t,e):void 0;return{mergedClsPrefix:n,mergedStyle:w(()=>{const{size:l,color:u}=e;return{fontSize:Ze(l),color:u}}),cssVars:i?void 0:t,themeClass:o==null?void 0:o.themeClass,onRender:o==null?void 0:o.onRender}},render(){var e;const{$parent:n,depth:i,mergedClsPrefix:r,component:t,onRender:o,themeClass:l}=this;return!((e=n==null?void 0:n.$options)===null||e===void 0)&&e._n_icon__&&ve("icon","don't wrap `n-icon` inside `n-icon`"),o==null||o(),s("i",le(this.$attrs,{role:"img",class:[`${r}-icon`,l,{[`${r}-icon--depth`]:i,[`${r}-icon--color-transition`]:i!==void 0}],style:[this.cssVars,this.mergedStyle]}),t?s(t):this.$slots)}});function ie(e,n){return e.type==="submenu"||e.type===void 0&&e[n]!==void 0}function po(e){return e.type==="group"}function ge(e){return e.type==="divider"}function fo(e){return e.type==="render"}const xe=_({name:"DropdownOption",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0},parentKey:{type:[String,Number],default:null},placement:{type:String,default:"right-start"},props:Object,scrollable:Boolean},setup(e){const n=H(J),{hoverKeyRef:i,keyboardKeyRef:r,lastToggledSubmenuKeyRef:t,pendingKeyPathRef:o,activeKeyPathRef:l,animatedRef:u,mergedShowRef:a,renderLabelRef:b,renderIconRef:v,labelFieldRef:g,childrenFieldRef:x,renderOptionRef:C,nodePropsRef:P,menuPropsRef:I}=n,S=H(pe,null),R=H(se),D=H(we),E=w(()=>e.tmNode.rawNode),W=w(()=>{const{value:d}=x;return ie(e.tmNode.rawNode,d)}),Q=w(()=>{const{disabled:d}=e.tmNode;return d}),Y=w(()=>{if(!W.value)return!1;const{key:d,disabled:f}=e.tmNode;if(f)return!1;const{value:y}=i,{value:B}=r,{value:te}=t,{value:A}=o;return y!==null?A.includes(d):B!==null?A.includes(d)&&A[A.length-1]!==d:te!==null?A.includes(d):!1}),ee=w(()=>r.value===null&&!u.value),oe=eo(Y,300,ee),ne=w(()=>!!(S!=null&&S.enteringSubmenuRef.value)),L=T(!1);j(pe,{enteringSubmenuRef:L});function M(){L.value=!0}function U(){L.value=!1}function z(){const{parentKey:d,tmNode:f}=e;f.disabled||a.value&&(t.value=d,r.value=null,i.value=f.key)}function c(){const{tmNode:d}=e;d.disabled||a.value&&i.value!==d.key&&z()}function p(d){if(e.tmNode.disabled||!a.value)return;const{relatedTarget:f}=d;f&&!ce({target:f},"dropdownOption")&&!ce({target:f},"scrollbarRail")&&(i.value=null)}function h(){const{value:d}=W,{tmNode:f}=e;a.value&&!d&&!f.disabled&&(n.doSelect(f.key,f.rawNode),n.doUpdateShow(!1))}return{labelField:g,renderLabel:b,renderIcon:v,siblingHasIcon:R.showIconRef,siblingHasSubmenu:R.hasSubmenuRef,menuProps:I,popoverBody:D,animated:u,mergedShowSubmenu:w(()=>oe.value&&!ne.value),rawNode:E,hasSubmenu:W,pending:q(()=>{const{value:d}=o,{key:f}=e.tmNode;return d.includes(f)}),childActive:q(()=>{const{value:d}=l,{key:f}=e.tmNode,y=d.findIndex(B=>f===B);return y===-1?!1:y<d.length-1}),active:q(()=>{const{value:d}=l,{key:f}=e.tmNode,y=d.findIndex(B=>f===B);return y===-1?!1:y===d.length-1}),mergedDisabled:Q,renderOption:C,nodeProps:P,handleClick:h,handleMouseMove:c,handleMouseEnter:z,handleMouseLeave:p,handleSubmenuBeforeEnter:M,handleSubmenuAfterEnter:U}},render(){var e,n;const{animated:i,rawNode:r,mergedShowSubmenu:t,clsPrefix:o,siblingHasIcon:l,siblingHasSubmenu:u,renderLabel:a,renderIcon:b,renderOption:v,nodeProps:g,props:x,scrollable:C}=this;let P=null;if(t){const D=(e=this.menuProps)===null||e===void 0?void 0:e.call(this,r,r.children);P=s(Se,Object.assign({},D,{clsPrefix:o,scrollable:this.scrollable,tmNodes:this.tmNode.children,parentKey:this.tmNode.key}))}const I={class:[`${o}-dropdown-option-body`,this.pending&&`${o}-dropdown-option-body--pending`,this.active&&`${o}-dropdown-option-body--active`,this.childActive&&`${o}-dropdown-option-body--child-active`,this.mergedDisabled&&`${o}-dropdown-option-body--disabled`],onMousemove:this.handleMouseMove,onMouseenter:this.handleMouseEnter,onMouseleave:this.handleMouseLeave,onClick:this.handleClick},S=g==null?void 0:g(r),R=s("div",Object.assign({class:[`${o}-dropdown-option`,S==null?void 0:S.class],"data-dropdown-option":!0},S),s("div",le(I,x),[s("div",{class:[`${o}-dropdown-option-body__prefix`,l&&`${o}-dropdown-option-body__prefix--show-icon`]},[b?b(r):X(r.icon)]),s("div",{"data-dropdown-option":!0,class:`${o}-dropdown-option-body__label`},a?a(r):X((n=r[this.labelField])!==null&&n!==void 0?n:r.title)),s("div",{"data-dropdown-option":!0,class:[`${o}-dropdown-option-body__suffix`,u&&`${o}-dropdown-option-body__suffix--has-submenu`]},this.hasSubmenu?s(uo,null,{default:()=>s(oo,null)}):null)]),this.hasSubmenu?s(Ke,null,{default:()=>[s(_e,null,{default:()=>s("div",{class:`${o}-dropdown-offset-container`},s(De,{show:this.mergedShowSubmenu,placement:this.placement,to:C&&this.popoverBody||void 0,teleportDisabled:!C},{default:()=>s("div",{class:`${o}-dropdown-menu-wrapper`},i?s(We,{onBeforeEnter:this.handleSubmenuBeforeEnter,onAfterEnter:this.handleSubmenuAfterEnter,name:"fade-in-scale-up-transition",appear:!0},{default:()=>P}):P)}))})]}):null);return v?v({node:R,option:r}):R}}),ho=_({name:"NDropdownGroup",props:{clsPrefix:{type:String,required:!0},tmNode:{type:Object,required:!0},parentKey:{type:[String,Number],default:null}},render(){const{tmNode:e,parentKey:n,clsPrefix:i}=this,{children:r}=e;return s(Ee,null,s(io,{clsPrefix:i,tmNode:e,key:e.key}),r==null?void 0:r.map(t=>{const{rawNode:o}=t;return o.show===!1?null:ge(o)?s(ye,{clsPrefix:i,key:t.key}):t.isGroup?(ve("dropdown","`group` node is not allowed to be put in `group` node."),null):s(xe,{clsPrefix:i,tmNode:t,parentKey:n,key:t.key})}))}}),vo=_({name:"DropdownRenderOption",props:{tmNode:{type:Object,required:!0}},render(){const{rawNode:{render:e,props:n}}=this.tmNode;return s("div",n,[e==null?void 0:e()])}}),Se=_({name:"DropdownMenu",props:{scrollable:Boolean,showArrow:Boolean,arrowStyle:[String,Object],clsPrefix:{type:String,required:!0},tmNodes:{type:Array,default:()=>[]},parentKey:{type:[String,Number],default:null}},setup(e){const{renderIconRef:n,childrenFieldRef:i}=H(J);j(se,{showIconRef:w(()=>{const t=n.value;return e.tmNodes.some(o=>{var l;if(o.isGroup)return(l=o.children)===null||l===void 0?void 0:l.some(({rawNode:a})=>t?t(a):a.icon);const{rawNode:u}=o;return t?t(u):u.icon})}),hasSubmenuRef:w(()=>{const{value:t}=i;return e.tmNodes.some(o=>{var l;if(o.isGroup)return(l=o.children)===null||l===void 0?void 0:l.some(({rawNode:a})=>ie(a,t));const{rawNode:u}=o;return ie(u,t)})})});const r=T(null);return j(Ve,null),j(Ge,null),j(we,r),{bodyRef:r}},render(){const{parentKey:e,clsPrefix:n,scrollable:i}=this,r=this.tmNodes.map(t=>{const{rawNode:o}=t;return o.show===!1?null:fo(o)?s(vo,{tmNode:t,key:t.key}):ge(o)?s(ye,{clsPrefix:n,key:t.key}):po(o)?s(ho,{clsPrefix:n,tmNode:t,parentKey:e,key:t.key}):s(xe,{clsPrefix:n,tmNode:t,parentKey:e,key:t.key,props:o.props,scrollable:i})});return s("div",{class:[`${n}-dropdown-menu`,i&&`${n}-dropdown-menu--scrollable`],ref:"bodyRef"},i?s(Ue,{contentClass:`${n}-dropdown-menu__content`},{default:()=>r}):r,this.showArrow?Oe({clsPrefix:n,arrowStyle:this.arrowStyle,arrowClass:void 0,arrowWrapperClass:void 0,arrowWrapperStyle:void 0}):null)}}),mo=k("dropdown-menu",`
 transform-origin: var(--v-transform-origin);
 background-color: var(--n-color);
 border-radius: var(--n-border-radius);
 box-shadow: var(--n-box-shadow);
 position: relative;
 transition:
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
`,[qe(),k("dropdown-option",`
 position: relative;
 `,[$("a",`
 text-decoration: none;
 color: inherit;
 outline: none;
 `,[$("&::before",`
 content: "";
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),k("dropdown-option-body",`
 display: flex;
 cursor: pointer;
 position: relative;
 height: var(--n-option-height);
 line-height: var(--n-option-height);
 font-size: var(--n-font-size);
 color: var(--n-option-text-color);
 transition: color .3s var(--n-bezier);
 `,[$("&::before",`
 content: "";
 position: absolute;
 top: 0;
 bottom: 0;
 left: 4px;
 right: 4px;
 transition: background-color .3s var(--n-bezier);
 border-radius: var(--n-border-radius);
 `),ue("disabled",[N("pending",`
 color: var(--n-option-text-color-hover);
 `,[O("prefix, suffix",`
 color: var(--n-option-text-color-hover);
 `),$("&::before","background-color: var(--n-option-color-hover);")]),N("active",`
 color: var(--n-option-text-color-active);
 `,[O("prefix, suffix",`
 color: var(--n-option-text-color-active);
 `),$("&::before","background-color: var(--n-option-color-active);")]),N("child-active",`
 color: var(--n-option-text-color-child-active);
 `,[O("prefix, suffix",`
 color: var(--n-option-text-color-child-active);
 `)])]),N("disabled",`
 cursor: not-allowed;
 opacity: var(--n-option-opacity-disabled);
 `),N("group",`
 font-size: calc(var(--n-font-size) - 1px);
 color: var(--n-group-header-text-color);
 `,[O("prefix",`
 width: calc(var(--n-option-prefix-width) / 2);
 `,[N("show-icon",`
 width: calc(var(--n-option-icon-prefix-width) / 2);
 `)])]),O("prefix",`
 width: var(--n-option-prefix-width);
 display: flex;
 justify-content: center;
 align-items: center;
 color: var(--n-prefix-color);
 transition: color .3s var(--n-bezier);
 z-index: 1;
 `,[N("show-icon",`
 width: var(--n-option-icon-prefix-width);
 `),k("icon",`
 font-size: var(--n-option-icon-size);
 `)]),O("label",`
 white-space: nowrap;
 flex: 1;
 z-index: 1;
 `),O("suffix",`
 box-sizing: border-box;
 flex-grow: 0;
 flex-shrink: 0;
 display: flex;
 justify-content: flex-end;
 align-items: center;
 min-width: var(--n-option-suffix-width);
 padding: 0 8px;
 transition: color .3s var(--n-bezier);
 color: var(--n-suffix-color);
 z-index: 1;
 `,[N("has-submenu",`
 width: var(--n-option-icon-suffix-width);
 `),k("icon",`
 font-size: var(--n-option-icon-size);
 `)]),k("dropdown-menu","pointer-events: all;")]),k("dropdown-offset-container",`
 pointer-events: none;
 position: absolute;
 left: 0;
 right: 0;
 top: -4px;
 bottom: -4px;
 `)]),k("dropdown-divider",`
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-divider-color);
 height: 1px;
 margin: 4px 0;
 `),k("dropdown-menu-wrapper",`
 transform-origin: var(--v-transform-origin);
 width: fit-content;
 `),$(">",[k("scrollbar",`
 height: inherit;
 max-height: inherit;
 `)]),ue("scrollable",`
 padding: var(--n-padding);
 `),N("scrollable",[O("content",`
 padding: var(--n-padding);
 `)])]),bo={animated:{type:Boolean,default:!0},keyboard:{type:Boolean,default:!0},size:String,inverted:Boolean,placement:{type:String,default:"bottom"},onSelect:[Function,Array],options:{type:Array,default:()=>[]},menuProps:Function,showArrow:Boolean,renderLabel:Function,renderIcon:Function,renderOption:Function,nodeProps:Function,labelField:{type:String,default:"label"},keyField:{type:String,default:"key"},childrenField:{type:String,default:"children"},value:[String,Number]},wo=Object.keys(fe),yo=Object.assign(Object.assign(Object.assign({},fe),bo),Z.props),ko=_({name:"Dropdown",inheritAttrs:!1,props:yo,setup(e){const n=T(!1),i=Je(K(e,"show"),n),r=w(()=>{const{keyField:c,childrenField:p}=e;return Be(e.options,{getKey(h){return h[c]},getDisabled(h){return h.disabled===!0},getIgnored(h){return h.type==="divider"||h.type==="render"},getChildren(h){return h[p]}})}),t=w(()=>r.value.treeNodes),o=T(null),l=T(null),u=T(null),a=w(()=>{var c,p,h;return(h=(p=(c=o.value)!==null&&c!==void 0?c:l.value)!==null&&p!==void 0?p:u.value)!==null&&h!==void 0?h:null}),b=w(()=>r.value.getPath(a.value).keyPath),v=w(()=>r.value.getPath(e.value).keyPath),g=q(()=>e.keyboard&&i.value);Ye({keydown:{ArrowUp:{prevent:!0,handler:ee},ArrowRight:{prevent:!0,handler:Y},ArrowDown:{prevent:!0,handler:oe},ArrowLeft:{prevent:!0,handler:Q},Enter:{prevent:!0,handler:ne},Escape:W}},g);const{mergedClsPrefixRef:x,inlineThemeDisabled:C,mergedComponentPropsRef:P}=me(e),I=w(()=>{var c,p;return e.size||((p=(c=P==null?void 0:P.value)===null||c===void 0?void 0:c.Dropdown)===null||p===void 0?void 0:p.size)||"medium"}),S=Z("Dropdown","-dropdown",mo,ro,e,x);j(J,{labelFieldRef:K(e,"labelField"),childrenFieldRef:K(e,"childrenField"),renderLabelRef:K(e,"renderLabel"),renderIconRef:K(e,"renderIcon"),hoverKeyRef:o,keyboardKeyRef:l,lastToggledSubmenuKeyRef:u,pendingKeyPathRef:b,activeKeyPathRef:v,animatedRef:K(e,"animated"),mergedShowRef:i,nodePropsRef:K(e,"nodeProps"),renderOptionRef:K(e,"renderOption"),menuPropsRef:K(e,"menuProps"),doSelect:R,doUpdateShow:D}),de(i,c=>{!e.animated&&!c&&E()});function R(c,p){const{onSelect:h}=e;h&&re(h,c,p)}function D(c){const{"onUpdate:show":p,onUpdateShow:h}=e;p&&re(p,c),h&&re(h,c),n.value=c}function E(){o.value=null,l.value=null,u.value=null}function W(){D(!1)}function Q(){M("left")}function Y(){M("right")}function ee(){M("up")}function oe(){M("down")}function ne(){const c=L();c!=null&&c.isLeaf&&i.value&&(R(c.key,c.rawNode),D(!1))}function L(){var c;const{value:p}=r,{value:h}=a;return!p||h===null?null:(c=p.getNode(h))!==null&&c!==void 0?c:null}function M(c){const{value:p}=a,{value:{getFirstAvailableNode:h}}=r;let d=null;if(p===null){const f=h();f!==null&&(d=f.key)}else{const f=L();if(f){let y;switch(c){case"down":y=f.getNext();break;case"up":y=f.getPrev();break;case"right":y=f.getChild();break;case"left":y=f.getParent();break}y&&(d=y.key)}}d!==null&&(o.value=null,l.value=d)}const U=w(()=>{const{inverted:c}=e,p=I.value,{common:{cubicBezierEaseInOut:h},self:d}=S.value,{padding:f,dividerColor:y,borderRadius:B,optionOpacityDisabled:te,[F("optionIconSuffixWidth",p)]:A,[F("optionSuffixWidth",p)]:Ce,[F("optionIconPrefixWidth",p)]:Pe,[F("optionPrefixWidth",p)]:ke,[F("fontSize",p)]:Ne,[F("optionHeight",p)]:Re,[F("optionIconSize",p)]:Ie}=d,m={"--n-bezier":h,"--n-font-size":Ne,"--n-padding":f,"--n-border-radius":B,"--n-option-height":Re,"--n-option-prefix-width":ke,"--n-option-icon-prefix-width":Pe,"--n-option-suffix-width":Ce,"--n-option-icon-suffix-width":A,"--n-option-icon-size":Ie,"--n-divider-color":y,"--n-option-opacity-disabled":te};return c?(m["--n-color"]=d.colorInverted,m["--n-option-color-hover"]=d.optionColorHoverInverted,m["--n-option-color-active"]=d.optionColorActiveInverted,m["--n-option-text-color"]=d.optionTextColorInverted,m["--n-option-text-color-hover"]=d.optionTextColorHoverInverted,m["--n-option-text-color-active"]=d.optionTextColorActiveInverted,m["--n-option-text-color-child-active"]=d.optionTextColorChildActiveInverted,m["--n-prefix-color"]=d.prefixColorInverted,m["--n-suffix-color"]=d.suffixColorInverted,m["--n-group-header-text-color"]=d.groupHeaderTextColorInverted):(m["--n-color"]=d.color,m["--n-option-color-hover"]=d.optionColorHover,m["--n-option-color-active"]=d.optionColorActive,m["--n-option-text-color"]=d.optionTextColor,m["--n-option-text-color-hover"]=d.optionTextColorHover,m["--n-option-text-color-active"]=d.optionTextColorActive,m["--n-option-text-color-child-active"]=d.optionTextColorChildActive,m["--n-prefix-color"]=d.prefixColor,m["--n-suffix-color"]=d.suffixColor,m["--n-group-header-text-color"]=d.groupHeaderTextColor),m}),z=C?be("dropdown",w(()=>`${I.value[0]}${e.inverted?"i":""}`),U,e):void 0;return{mergedClsPrefix:x,mergedTheme:S,mergedSize:I,tmNodes:t,mergedShow:i,handleAfterLeave:()=>{e.animated&&E()},doUpdateShow:D,cssVars:C?void 0:U,themeClass:z==null?void 0:z.themeClass,onRender:z==null?void 0:z.onRender}},render(){const e=(r,t,o,l,u)=>{var a;const{mergedClsPrefix:b,menuProps:v}=this;(a=this.onRender)===null||a===void 0||a.call(this);const g=(v==null?void 0:v(void 0,this.tmNodes.map(C=>C.rawNode)))||{},x={ref:Qe(t),class:[r,`${b}-dropdown`,`${b}-dropdown--${this.mergedSize}-size`,this.themeClass],clsPrefix:b,tmNodes:this.tmNodes,style:[...o,this.cssVars],showArrow:this.showArrow,arrowStyle:this.arrowStyle,scrollable:this.scrollable,onMouseenter:l,onMouseleave:u};return s(Se,le(this.$attrs,x,g))},{mergedTheme:n}=this,i={show:this.mergedShow,theme:n.peers.Popover,themeOverrides:n.peerOverrides.Popover,internalOnAfterLeave:this.handleAfterLeave,internalRenderBody:e,onUpdateShow:this.doUpdateShow,"onUpdate:show":void 0};return s($e,Object.assign({},Xe(this.$props,wo),i),{trigger:()=>{var r,t;return(t=(r=this.$slots).default)===null||t===void 0?void 0:t.call(r)}})}});export{ko as N};
