import{G as _,cb as t,K as o,L as C,M as v,aH as N,P as F,Q as K,R as D,J as G,W as I,Z as z,_ as J,$ as Q,v as f,a3 as x}from"./index-BNsw-iSP.js";const Z={thPaddingSmall:"6px",thPaddingMedium:"12px",thPaddingLarge:"12px",tdPaddingSmall:"6px",tdPaddingMedium:"12px",tdPaddingLarge:"12px"};function q(e){const{dividerColor:r,cardColor:a,modalColor:s,popoverColor:l,tableHeaderColor:c,tableColorStriped:b,textColor1:h,textColor2:g,borderRadius:n,fontWeightStrong:d,lineHeight:i,fontSizeSmall:p,fontSizeMedium:m,fontSizeLarge:u}=e;return Object.assign(Object.assign({},Z),{fontSizeSmall:p,fontSizeMedium:m,fontSizeLarge:u,lineHeight:i,borderRadius:n,borderColor:t(a,r),borderColorModal:t(s,r),borderColorPopover:t(l,r),tdColor:a,tdColorModal:s,tdColorPopover:l,tdColorStriped:t(a,b),tdColorStripedModal:t(s,b),tdColorStripedPopover:t(l,b),thColor:t(a,c),thColorModal:t(s,c),thColorPopover:t(l,c),thTextColor:h,tdTextColor:g,thFontWeight:d})}const A={common:_,self:q},U=o([C("table",`
 font-size: var(--n-font-size);
 font-variant-numeric: tabular-nums;
 line-height: var(--n-line-height);
 width: 100%;
 border-radius: var(--n-border-radius) var(--n-border-radius) 0 0;
 text-align: left;
 border-collapse: separate;
 border-spacing: 0;
 overflow: hidden;
 background-color: var(--n-td-color);
 border-color: var(--n-merged-border-color);
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 --n-merged-border-color: var(--n-border-color);
 `,[o("th",`
 white-space: nowrap;
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 text-align: inherit;
 padding: var(--n-th-padding);
 vertical-align: inherit;
 text-transform: none;
 border: 0px solid var(--n-merged-border-color);
 font-weight: var(--n-th-font-weight);
 color: var(--n-th-text-color);
 background-color: var(--n-th-color);
 border-bottom: 1px solid var(--n-merged-border-color);
 border-right: 1px solid var(--n-merged-border-color);
 `,[o("&:last-child",`
 border-right: 0px solid var(--n-merged-border-color);
 `)]),o("td",`
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 padding: var(--n-td-padding);
 color: var(--n-td-text-color);
 background-color: var(--n-td-color);
 border: 0px solid var(--n-merged-border-color);
 border-right: 1px solid var(--n-merged-border-color);
 border-bottom: 1px solid var(--n-merged-border-color);
 `,[o("&:last-child",`
 border-right: 0px solid var(--n-merged-border-color);
 `)]),v("bordered",`
 border: 1px solid var(--n-merged-border-color);
 border-radius: var(--n-border-radius);
 `,[o("tr",[o("&:last-child",[o("td",`
 border-bottom: 0 solid var(--n-merged-border-color);
 `)])])]),v("single-line",[o("th",`
 border-right: 0px solid var(--n-merged-border-color);
 `),o("td",`
 border-right: 0px solid var(--n-merged-border-color);
 `)]),v("single-column",[o("tr",[o("&:not(:last-child)",[o("td",`
 border-bottom: 0px solid var(--n-merged-border-color);
 `)])])]),v("striped",[o("tr:nth-of-type(even)",[o("td","background-color: var(--n-td-color-striped)")])]),N("bottom-bordered",[o("tr",[o("&:last-child",[o("td",`
 border-bottom: 0px solid var(--n-merged-border-color);
 `)])])])]),F(C("table",`
 background-color: var(--n-td-color-modal);
 --n-merged-border-color: var(--n-border-color-modal);
 `,[o("th",`
 background-color: var(--n-th-color-modal);
 `),o("td",`
 background-color: var(--n-td-color-modal);
 `)])),K(C("table",`
 background-color: var(--n-td-color-popover);
 --n-merged-border-color: var(--n-border-color-popover);
 `,[o("th",`
 background-color: var(--n-th-color-popover);
 `),o("td",`
 background-color: var(--n-td-color-popover);
 `)]))]),X=Object.assign(Object.assign({},z.props),{bordered:{type:Boolean,default:!0},bottomBordered:{type:Boolean,default:!0},singleLine:{type:Boolean,default:!0},striped:Boolean,singleColumn:Boolean,size:String}),oo=D({name:"Table",props:X,setup(e){const{mergedClsPrefixRef:r,inlineThemeDisabled:a,mergedRtlRef:s,mergedComponentPropsRef:l}=I(e),c=f(()=>{var d,i;return e.size||((i=(d=l==null?void 0:l.value)===null||d===void 0?void 0:d.Table)===null||i===void 0?void 0:i.size)||"medium"}),b=z("Table","-table",U,A,e,r),h=J("Table",s,r),g=f(()=>{const d=c.value,{self:{borderColor:i,tdColor:p,tdColorModal:m,tdColorPopover:u,thColor:P,thColorModal:S,thColorPopover:M,thTextColor:R,tdTextColor:k,borderRadius:T,thFontWeight:B,lineHeight:$,borderColorModal:w,borderColorPopover:y,tdColorStriped:L,tdColorStripedModal:H,tdColorStripedPopover:O,[x("fontSize",d)]:j,[x("tdPadding",d)]:E,[x("thPadding",d)]:V},common:{cubicBezierEaseInOut:W}}=b.value;return{"--n-bezier":W,"--n-td-color":p,"--n-td-color-modal":m,"--n-td-color-popover":u,"--n-td-text-color":k,"--n-border-color":i,"--n-border-color-modal":w,"--n-border-color-popover":y,"--n-border-radius":T,"--n-font-size":j,"--n-th-color":P,"--n-th-color-modal":S,"--n-th-color-popover":M,"--n-th-font-weight":B,"--n-th-text-color":R,"--n-line-height":$,"--n-td-padding":E,"--n-th-padding":V,"--n-td-color-striped":L,"--n-td-color-striped-modal":H,"--n-td-color-striped-popover":O}}),n=a?Q("table",f(()=>c.value[0]),g,e):void 0;return{rtlEnabled:h,mergedClsPrefix:r,cssVars:a?void 0:g,themeClass:n==null?void 0:n.themeClass,onRender:n==null?void 0:n.onRender}},render(){var e;const{mergedClsPrefix:r}=this;return(e=this.onRender)===null||e===void 0||e.call(this),G("table",{class:[`${r}-table`,this.themeClass,{[`${r}-table--rtl`]:this.rtlEnabled,[`${r}-table--bottom-bordered`]:this.bottomBordered,[`${r}-table--bordered`]:this.bordered,[`${r}-table--single-line`]:this.singleLine,[`${r}-table--single-column`]:this.singleColumn,[`${r}-table--striped`]:this.striped}],style:this.cssVars},this.$slots)}});export{oo as N};
