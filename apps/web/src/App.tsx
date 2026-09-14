import { FormEvent, useEffect, useMemo, useState } from 'react'
import { BarChart3, Building2, CreditCard, Download, FileText, LayoutDashboard, LogOut, Package, Plus, ReceiptText, Search, Settings, ShieldCheck, Trash2, Users, WalletCards, X } from 'lucide-react'

type Company = { id:string; name:string; slug:string; timezone:string; currency:string; role:string }
type User = { id:string; email:string; display_name:string }
type Me = { user:User; company?:Company }
type Customer = { id:string; name:string; email?:string; phone?:string; tax_id?:string; address?:string; notes?:string; created_at?:string }
type CustomerInput = { name:string; email:string; phone:string; tax_id:string; address:string; notes:string }
type Product = { id:string; type:'product'|'service'; name:string; sku?:string; description?:string; unit:string; price_minor:number; tax_rate_bps:number; is_active:boolean }
type Quote = { id:string; number:string; status:string; customer_id:string; customer_name:string; issue_date:string; valid_until?:string; subtotal_minor:number; discount_minor:number; tax_minor:number; total_minor:number; notes?:string }
type Invoice = { id:string; number:string; status:string; customer_id:string; customer_name:string; issue_date:string; due_date?:string; total_minor:number; paid_minor:number; balance_minor:number; quotation_id?:string }
type Payment = { id:string; invoice_id:string; invoice_number:string; amount_minor:number; method:string; reference?:string; paid_at:string }
type Expense = { id:string; category:string; expense_date:string; vendor:string; description:string; amount_minor:number; payment_method:string; reference:string }
type FinanceReport = { invoiced_minor:number; collected_minor:number; expense_minor:number; cash_result_minor:number; receivable_minor:number; open_invoices:number }
type BusinessProfile = { legal_name:string; tax_id:string; email:string; phone:string; address:string; bank_name:string; bank_account_name:string; bank_account_number:string; qris_text:string }
type Page = 'overview'|'customers'|'products'|'quotations'|'invoices'|'payments'|'expenses'|'reports'|'settings'

const emptyCustomer: CustomerInput = { name:'', email:'', phone:'', tax_id:'', address:'', notes:'' }

async function api<T>(url:string, init:RequestInit = {}):Promise<T>{
  const res = await fetch(url,{credentials:'include',...init,headers:{'Content-Type':'application/json',...(init.headers||{})}})
  if(res.status===204) return undefined as T
  const body = await res.json().catch(()=>({error:'Unexpected server response'}))
  if(!res.ok) throw new Error(body.error || 'Request failed')
  return body.data as T
}

export default function App(){
  const [me,setMe]=useState<Me|null>(null)
  const [loading,setLoading]=useState(true)
  const [authMode,setAuthMode]=useState<'login'|'register'>('login')
  const [page,setPage]=useState<Page>('overview')

  async function refreshMe(){
    try{ setMe(await api<Me>('/api/v1/me')) }catch{ setMe(null) }finally{ setLoading(false) }
  }
  useEffect(()=>{ refreshMe() },[])

  if(loading) return <Splash/>
  if(!me) return <AuthScreen mode={authMode} setMode={setAuthMode} onDone={refreshMe}/>
  if(!me.company) return <Onboarding me={me} onDone={refreshMe}/>
  return <Shell me={me} page={page} setPage={setPage} onLogout={async()=>{await api('/api/v1/auth/logout',{method:'POST'});setMe(null)}}/>
}

function Splash(){ return <div className="center-screen"><div className="logo-orb">V</div><p>Starting VENOMBusiness…</p></div> }

function AuthScreen({mode,setMode,onDone}:{mode:'login'|'register';setMode:(m:'login'|'register')=>void;onDone:()=>Promise<void>}){
  const [displayName,setDisplayName]=useState('')
  const [email,setEmail]=useState('')
  const [password,setPassword]=useState('')
  const [error,setError]=useState('')
  const [busy,setBusy]=useState(false)
  async function submit(e:FormEvent){
    e.preventDefault(); setBusy(true); setError('')
    try{
      await api(`/api/v1/auth/${mode}`,{method:'POST',body:JSON.stringify(mode==='register'?{display_name:displayName,email,password}:{email,password})})
      await onDone()
    }catch(e){setError(e instanceof Error?e.message:'Request failed')}finally{setBusy(false)}
  }
  return <div className="auth-layout">
    <section className="auth-hero">
      <div className="brand big"><span className="mark">V</span><div><b>VENOM</b><small>BUSINESS</small></div></div>
      <div className="hero-copy"><span className="eyebrow">OPEN-SOURCE BUSINESS OS</span><h1>Run the business.<br/>Not the software.</h1><p>One clean workspace for customers, finance, sales and operations—built for small teams that want control without enterprise complexity.</p></div>
      <div className="security-note"><ShieldCheck size={18}/><span>Secure session auth · Argon2id · Role-based access</span></div>
    </section>
    <section className="auth-card-wrap"><form className="auth-card" onSubmit={submit}>
      <small>{mode==='login'?'WELCOME BACK':'CREATE YOUR WORKSPACE'}</small>
      <h2>{mode==='login'?'Sign in to VENOMBusiness':'Start with your account'}</h2>
      <p>{mode==='login'?'Continue managing your business workspace.':'You will set up your business right after this step.'}</p>
      {mode==='register'&&<label>Full name<input value={displayName} onChange={e=>setDisplayName(e.target.value)} placeholder="John Doe" required minLength={2}/></label>}
      <label>Email<input value={email} onChange={e=>setEmail(e.target.value)} type="email" placeholder="you@company.com" required/></label>
      <label>Password<input value={password} onChange={e=>setPassword(e.target.value)} type="password" placeholder="Minimum 10 characters" required minLength={10}/></label>
      {error&&<div className="error-box">{error}</div>}
      <button className="primary full" disabled={busy}>{busy?'Please wait…':mode==='login'?'Sign in':'Create account'}</button>
      <button type="button" className="text-button" onClick={()=>setMode(mode==='login'?'register':'login')}>{mode==='login'?'New here? Create an account':'Already have an account? Sign in'}</button>
    </form></section>
  </div>
}

function Onboarding({me,onDone}:{me:Me;onDone:()=>Promise<void>}){
  const [name,setName]=useState('')
  const [slug,setSlug]=useState('')
  const [currency,setCurrency]=useState('IDR')
  const [timezone,setTimezone]=useState('Asia/Jakarta')
  const [error,setError]=useState('')
  async function submit(e:FormEvent){e.preventDefault();setError('');try{await api('/api/v1/company/setup',{method:'POST',body:JSON.stringify({name,slug,currency,timezone})});await onDone()}catch(e){setError(e instanceof Error?e.message:'Could not create workspace')}}
  return <div className="onboarding"><div className="onboard-card">
    <div className="progress"><i className="done"/><i className="active"/><i/></div><span className="eyebrow">BUSINESS SETUP</span>
    <h1>Welcome, {me.user.display_name.split(' ')[0]}.</h1><p>Tell us the basics. These become the defaults for invoices, reports and future modules.</p>
    <form onSubmit={submit} className="setup-grid">
      <label className="wide">Business name<input value={name} onChange={e=>{setName(e.target.value);if(!slug)setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,''))}} placeholder="Acme Studio" required/></label>
      <label className="wide">Workspace slug<input value={slug} onChange={e=>setSlug(e.target.value)} placeholder="acme-studio"/></label>
      <label>Currency<select value={currency} onChange={e=>setCurrency(e.target.value)}><option>IDR</option><option>USD</option><option>SGD</option><option>MYR</option><option>THB</option></select></label>
      <label>Timezone<select value={timezone} onChange={e=>setTimezone(e.target.value)}><option>Asia/Jakarta</option><option>Asia/Bangkok</option><option>Asia/Singapore</option><option>UTC</option></select></label>
      {error&&<div className="error-box wide">{error}</div>}
      <button className="primary wide">Create business workspace</button>
    </form>
  </div></div>
}

function Shell({me,page,setPage,onLogout}:{me:Me;page:Page;setPage:(p:Page)=>void;onLogout:()=>void}){
  const nav = [
    {id:'overview' as Page,label:'Overview',icon:LayoutDashboard},
    {id:'customers' as Page,label:'Customers',icon:Users},
    {id:'products' as Page,label:'Products & Services',icon:Package},
    {id:'quotations' as Page,label:'Quotations',icon:FileText},
    {id:'invoices' as Page,label:'Invoices',icon:ReceiptText},
    {id:'payments' as Page,label:'Payments',icon:CreditCard},
    {id:'expenses' as Page,label:'Expenses',icon:WalletCards},
    {id:'reports' as Page,label:'Reports',icon:BarChart3},
    {id:'settings' as Page,label:'Business Profile',icon:Settings},
  ]
  let content = <Overview me={me}/>
  if(page==='customers') content=<Customers role={me.company!.role}/>
  if(page==='products') content=<Products role={me.company!.role} currency={me.company!.currency}/>
  if(page==='quotations') content=<Quotations role={me.company!.role} currency={me.company!.currency}/>
  if(page==='invoices') content=<Invoices role={me.company!.role} currency={me.company!.currency}/>
  if(page==='payments') content=<Payments role={me.company!.role} currency={me.company!.currency}/>
  if(page==='expenses') content=<Expenses role={me.company!.role} currency={me.company!.currency}/>
  if(page==='reports') content=<Reports currency={me.company!.currency}/>
  if(page==='settings') content=<SettingsPage me={me}/>
  return <div className="shell"><aside>
    <div className="brand"><span className="mark">V</span><div><b>VENOM</b><small>BUSINESS</small></div></div>
    <nav>{nav.map(n=><button className={page===n.id?'active':''} onClick={()=>setPage(n.id)} key={n.id}><n.icon size={18}/><span>{n.label}</span></button>)}</nav>
    <div className="workspace"><Building2 size={18}/><div><small>WORKSPACE</small><b>{me.company!.name}</b><em>{me.company!.role}</em></div></div>
    <button className="logout" onClick={onLogout}><LogOut size={17}/><span>Sign out</span></button>
  </aside><main>{content}</main></div>
}
function Overview({me}:{me:Me}){
 const [d,setD]=useState<any>({customers:0,products:0,open_invoices:0,invoiced_minor:0,paid_minor:0,outstanding_minor:0})
 useEffect(()=>{api<any>('/api/v1/dashboard').then(setD).catch(()=>{})},[])
 return <><header><div><p>VENOMBUSINESS V0.3</p><h1>Good to see you, {me.user.display_name.split(' ')[0]}.</h1><span>Your commercial workflow is now connected end-to-end.</span></div></header>
 <section className="cards"><article><small>CUSTOMERS</small><strong>{d.customers}</strong><span>Active directory</span></article><article><small>CATALOG</small><strong>{d.products}</strong><span>Active products & services</span></article><article><small>OUTSTANDING</small><strong>{money(d.outstanding_minor,me.company!.currency)}</strong><span>{d.open_invoices} open invoices</span></article><article><small>COLLECTED</small><strong>{money(d.paid_minor,me.company!.currency)}</strong><span>Total recorded payments</span></article></section>
 <section className="grid"><article className="panel roadmap"><small>COMMERCE MILESTONE</small><h2>v0.3 commerce core is active</h2><p>Catalog, quotations, invoice conversion and payment recording now share one company-scoped data model with audit trails.</p><div className="milestones"><span className="complete">Auth</span><span className="complete">Customers</span><span className="complete">Products</span><span className="complete">Quotation</span><span className="complete">Invoice</span><span className="complete">Payment</span><span>Expenses</span><span>Reports</span></div></article><article className="panel"><small>FINANCIAL SNAPSHOT</small><h2>{money(d.invoiced_minor,me.company!.currency)}</h2><p className="muted">Total non-void invoiced value. Money values are stored as integer minor units to avoid floating-point rounding issues.</p></article></section></>
}
function Customers({role}:{role:string}){
 const [items,setItems]=useState<Customer[]>([]), [query,setQuery]=useState(''), [modal,setModal]=useState(false), [editing,setEditing]=useState<Customer|null>(null), [form,setForm]=useState<CustomerInput>(emptyCustomer), [error,setError]=useState('')
 const canDelete=['owner','admin','manager'].includes(role)
 const filtered=useMemo(()=>items.filter(c=>`${c.name} ${c.email||''} ${c.phone||''}`.toLowerCase().includes(query.toLowerCase())),[items,query])
 async function load(){try{setItems(await api<Customer[]>('/api/v1/customers'))}catch(e){setError(e instanceof Error?e.message:'Could not load customers')}}
 useEffect(()=>{load()},[])
 function openCreate(){setEditing(null);setForm(emptyCustomer);setError('');setModal(true)}
 function openEdit(c:Customer){setEditing(c);setForm({name:c.name,email:c.email||'',phone:c.phone||'',tax_id:c.tax_id||'',address:c.address||'',notes:c.notes||''});setError('');setModal(true)}
 async function save(e:FormEvent){e.preventDefault();setError('');try{await api(editing?`/api/v1/customers/${editing.id}`:'/api/v1/customers',{method:editing?'PUT':'POST',body:JSON.stringify(form)});setModal(false);await load()}catch(e){setError(e instanceof Error?e.message:'Could not save customer')}}
 async function remove(c:Customer){if(!confirm(`Delete ${c.name}?`))return;try{await api(`/api/v1/customers/${c.id}`,{method:'DELETE'});await load()}catch(e){setError(e instanceof Error?e.message:'Could not delete customer')}}
 return <><header><div><p>CUSTOMER DIRECTORY</p><h1>Customers</h1><span>Every record is isolated inside your business workspace.</span></div><button className="primary" onClick={openCreate}><Plus size={17}/> New customer</button></header>
 <div className="toolbar"><div className="search"><Search size={17}/><input placeholder="Search name, email or phone…" value={query} onChange={e=>setQuery(e.target.value)}/></div><span>{filtered.length} customers</span></div>{error&&!modal&&<div className="error-box">{error}</div>}
 <div className="table-wrap"><table><thead><tr><th>Customer</th><th>Contact</th><th>Tax ID</th><th>Address</th><th></th></tr></thead><tbody>{filtered.length===0?<tr><td colSpan={5}><div className="empty"><Users size={28}/><b>No customers yet</b><span>Create your first customer to test the live CRUD flow.</span></div></td></tr>:filtered.map(c=><tr key={c.id}><td><b>{c.name}</b></td><td><span>{c.email||'—'}</span><small>{c.phone||''}</small></td><td>{c.tax_id||'—'}</td><td className="truncate">{c.address||'—'}</td><td><div className="row-actions"><button onClick={()=>openEdit(c)}>Edit</button>{canDelete&&<button className="danger" onClick={()=>remove(c)}><Trash2 size={15}/></button>}</div></td></tr>)}</tbody></table></div>
 {modal&&<div className="modal-backdrop" onMouseDown={()=>setModal(false)}><form className="modal" onSubmit={save} onMouseDown={e=>e.stopPropagation()}><div className="modal-head"><div><small>{editing?'EDIT RECORD':'NEW RECORD'}</small><h2>{editing?'Update customer':'Add customer'}</h2></div><button type="button" className="icon" onClick={()=>setModal(false)}><X/></button></div><div className="form-grid"><label className="wide">Customer / Company name<input required value={form.name} onChange={e=>setForm({...form,name:e.target.value})}/></label><label>Email<input type="email" value={form.email} onChange={e=>setForm({...form,email:e.target.value})}/></label><label>Phone<input value={form.phone} onChange={e=>setForm({...form,phone:e.target.value})}/></label><label>Tax ID / NPWP<input value={form.tax_id} onChange={e=>setForm({...form,tax_id:e.target.value})}/></label><label className="wide">Address<textarea value={form.address} onChange={e=>setForm({...form,address:e.target.value})}/></label><label className="wide">Notes<textarea value={form.notes} onChange={e=>setForm({...form,notes:e.target.value})}/></label></div>{error&&<div className="error-box">{error}</div>}<div className="modal-actions"><button type="button" onClick={()=>setModal(false)}>Cancel</button><button className="primary">{editing?'Save changes':'Create customer'}</button></div></form></div>}</>
}

function money(v:number,currency:string){return new Intl.NumberFormat('id-ID',{style:'currency',currency,maximumFractionDigits:currency==='IDR'?0:2}).format((v||0)/100)}
function toMinor(v:string){return Math.round((Number(v)||0)*100)}
function today(){return new Date().toISOString().slice(0,10)}

function Products({role,currency}:{role:string;currency:string}){
 const [items,setItems]=useState<Product[]>([]),[query,setQuery]=useState(''),[modal,setModal]=useState(false),[editing,setEditing]=useState<Product|null>(null),[error,setError]=useState('')
 const [form,setForm]=useState({type:'service' as 'product'|'service',name:'',sku:'',description:'',unit:'item',price:'',tax:'0',is_active:true})
 const canDelete=['owner','admin','manager'].includes(role)
 async function load(){try{setItems(await api<Product[]>('/api/v1/products'))}catch(e){setError(e instanceof Error?e.message:'Could not load catalog')}}
 useEffect(()=>{load()},[])
 const filtered=items.filter(x=>`${x.name} ${x.sku||''}`.toLowerCase().includes(query.toLowerCase()))
 function open(x?:Product){setEditing(x||null);setError('');setForm(x?{type:x.type,name:x.name,sku:x.sku||'',description:x.description||'',unit:x.unit,price:String(x.price_minor/100),tax:String(x.tax_rate_bps/100),is_active:x.is_active}:{type:'service',name:'',sku:'',description:'',unit:'item',price:'',tax:'0',is_active:true});setModal(true)}
 async function save(e:FormEvent){e.preventDefault();try{await api(editing?`/api/v1/products/${editing.id}`:'/api/v1/products',{method:editing?'PUT':'POST',body:JSON.stringify({type:form.type,name:form.name,sku:form.sku,description:form.description,unit:form.unit,price_minor:toMinor(form.price),tax_rate_bps:Math.round(Number(form.tax)*100),is_active:form.is_active})});setModal(false);await load()}catch(e){setError(e instanceof Error?e.message:'Could not save item')}}
 async function remove(x:Product){if(!confirm(`Delete ${x.name}?`))return;try{await api(`/api/v1/products/${x.id}`,{method:'DELETE'});await load()}catch(e){setError(e instanceof Error?e.message:'Could not delete item')}}
 return <><header><div><p>CATALOG</p><h1>Products & Services</h1><span>Reusable pricing data for quotations and invoices.</span></div><button className="primary" onClick={()=>open()}><Plus size={17}/> New item</button></header><div className="toolbar"><div className="search"><Search size={17}/><input placeholder="Search name or SKU…" value={query} onChange={e=>setQuery(e.target.value)}/></div><span>{filtered.length} items</span></div>{error&&!modal&&<div className="error-box">{error}</div>}<div className="table-wrap"><table><thead><tr><th>Item</th><th>Type</th><th>SKU</th><th>Price</th><th>Tax</th><th>Status</th><th></th></tr></thead><tbody>{filtered.length===0?<Empty cols={7} icon={<Package size={28}/>} title="No catalog items" text="Add a product or service to start creating quotations."/>:filtered.map(x=><tr key={x.id}><td><b>{x.name}</b><small>{x.unit}</small></td><td className="capitalize">{x.type}</td><td>{x.sku||'—'}</td><td>{money(x.price_minor,currency)}</td><td>{(x.tax_rate_bps/100).toFixed(2)}%</td><td><Status value={x.is_active?'active':'inactive'}/></td><td><div className="row-actions"><button onClick={()=>open(x)}>Edit</button>{canDelete&&<button className="danger" onClick={()=>remove(x)}><Trash2 size={15}/></button>}</div></td></tr>)}</tbody></table></div>
 {modal&&<div className="modal-backdrop" onMouseDown={()=>setModal(false)}><form className="modal" onSubmit={save} onMouseDown={e=>e.stopPropagation()}><div className="modal-head"><div><small>CATALOG ITEM</small><h2>{editing?'Update item':'Add product or service'}</h2></div><button type="button" className="icon" onClick={()=>setModal(false)}><X/></button></div><div className="form-grid"><label>Type<select value={form.type} onChange={e=>setForm({...form,type:e.target.value as any})}><option value="service">Service</option><option value="product">Product</option></select></label><label>SKU<input value={form.sku} onChange={e=>setForm({...form,sku:e.target.value})}/></label><label className="wide">Name<input required value={form.name} onChange={e=>setForm({...form,name:e.target.value})}/></label><label>Unit<input value={form.unit} onChange={e=>setForm({...form,unit:e.target.value})}/></label><label>Price ({currency})<input type="number" min="0" step="0.01" required value={form.price} onChange={e=>setForm({...form,price:e.target.value})}/></label><label>Tax %<input type="number" min="0" max="100" step="0.01" value={form.tax} onChange={e=>setForm({...form,tax:e.target.value})}/></label><label className="check"><input type="checkbox" checked={form.is_active} onChange={e=>setForm({...form,is_active:e.target.checked})}/> Active item</label><label className="wide">Description<textarea value={form.description} onChange={e=>setForm({...form,description:e.target.value})}/></label></div>{error&&<div className="error-box">{error}</div>}<div className="modal-actions"><button type="button" onClick={()=>setModal(false)}>Cancel</button><button className="primary">Save item</button></div></form></div>}</>
}

type QuoteLine={product_id:string;name:string;description:string;quantity:string;unit:string;price:string;tax:string}
function Quotations({role,currency}:{role:string;currency:string}){
 const [items,setItems]=useState<Quote[]>([]),[customers,setCustomers]=useState<Customer[]>([]),[products,setProducts]=useState<Product[]>([]),[modal,setModal]=useState(false),[error,setError]=useState('')
 const [customer,setCustomer]=useState(''),[valid,setValid]=useState(''),[discount,setDiscount]=useState('0'),[notes,setNotes]=useState(''),[lines,setLines]=useState<QuoteLine[]>([{product_id:'',name:'',description:'',quantity:'1',unit:'item',price:'',tax:'0'}])
 const canConvert=['owner','admin','manager'].includes(role)
 async function load(){try{const [q,c,p]=await Promise.all([api<Quote[]>('/api/v1/quotations'),api<Customer[]>('/api/v1/customers'),api<Product[]>('/api/v1/products')]);setItems(q);setCustomers(c);setProducts(p.filter(x=>x.is_active))}catch(e){setError(e instanceof Error?e.message:'Could not load quotations')}}
 useEffect(()=>{load()},[])
 function selectProduct(i:number,pid:string){const p=products.find(x=>x.id===pid);setLines(lines.map((l,n)=>n===i?(p?{product_id:p.id,name:p.name,description:p.description||'',quantity:l.quantity,unit:p.unit,price:String(p.price_minor/100),tax:String(p.tax_rate_bps/100)}:{...l,product_id:''}):l))}
 async function save(e:FormEvent){e.preventDefault();try{await api('/api/v1/quotations',{method:'POST',body:JSON.stringify({customer_id:customer,issue_date:today(),valid_until:valid,discount_minor:toMinor(discount),notes,items:lines.map(l=>({product_id:l.product_id,name:l.name,description:l.description,quantity_milli:Math.round(Number(l.quantity)*1000),unit:l.unit,unit_price_minor:toMinor(l.price),tax_rate_bps:Math.round(Number(l.tax)*100)}))})});setModal(false);setLines([{product_id:'',name:'',description:'',quantity:'1',unit:'item',price:'',tax:'0'}]);setCustomer('');await load()}catch(e){setError(e instanceof Error?e.message:'Could not create quotation')}}
 async function status(q:Quote,value:string){try{await api(`/api/v1/quotations/${q.id}/status`,{method:'PATCH',body:JSON.stringify({status:value})});await load()}catch(e){setError(e instanceof Error?e.message:'Could not update quotation')}}
 async function convert(q:Quote){if(!confirm(`Convert ${q.number} to invoice?`))return;try{await api(`/api/v1/quotations/${q.id}/invoice`,{method:'POST'});await load();alert('Invoice created successfully.')}catch(e){setError(e instanceof Error?e.message:'Could not create invoice')}}
 return <><header><div><p>SALES</p><h1>Quotations</h1><span>Send structured offers and convert accepted work into invoices.</span></div><button className="primary" onClick={()=>{setError('');setModal(true)}}><Plus size={17}/> New quotation</button></header>{error&&!modal&&<div className="error-box">{error}</div>}<div className="table-wrap"><table><thead><tr><th>Number</th><th>Customer</th><th>Issue</th><th>Total</th><th>Status</th><th></th></tr></thead><tbody>{items.length===0?<Empty cols={6} icon={<FileText size={28}/>} title="No quotations" text="Create a quotation from your customer and catalog data."/>:items.map(q=><tr key={q.id}><td><b>{q.number}</b></td><td>{q.customer_name}</td><td>{q.issue_date}</td><td>{money(q.total_minor,currency)}</td><td><select className="status-select" value={q.status} onChange={e=>status(q,e.target.value)}><option>draft</option><option>sent</option><option>accepted</option><option>rejected</option><option>expired</option></select></td><td><div className="row-actions"><a className="button-link" target="_blank" href={`/api/v1/documents/quotations/${q.id}/pdf`}><Download size={14}/> PDF</a>{canConvert&&q.status==='accepted'&&<button onClick={()=>convert(q)}>Create invoice</button>}</div></td></tr>)}</tbody></table></div>
 {modal&&<div className="modal-backdrop" onMouseDown={()=>setModal(false)}><form className="modal modal-wide" onSubmit={save} onMouseDown={e=>e.stopPropagation()}><div className="modal-head"><div><small>NEW SALES DOCUMENT</small><h2>Create quotation</h2></div><button type="button" className="icon" onClick={()=>setModal(false)}><X/></button></div><div className="form-grid"><label>Customer<select required value={customer} onChange={e=>setCustomer(e.target.value)}><option value="">Choose customer…</option>{customers.map(c=><option key={c.id} value={c.id}>{c.name}</option>)}</select></label><label>Valid until<input type="date" value={valid} onChange={e=>setValid(e.target.value)}/></label></div><div className="line-items"><div className="line-head"><b>Line items</b><button type="button" onClick={()=>setLines([...lines,{product_id:'',name:'',description:'',quantity:'1',unit:'item',price:'',tax:'0'}])}><Plus size={14}/> Add line</button></div>{lines.map((l,i)=><div className="line" key={i}><select value={l.product_id} onChange={e=>selectProduct(i,e.target.value)}><option value="">Custom item…</option>{products.map(p=><option key={p.id} value={p.id}>{p.name}</option>)}</select><input placeholder="Name" required value={l.name} onChange={e=>setLines(lines.map((x,n)=>n===i?{...x,name:e.target.value}:x))}/><input className="qty" type="number" min="0.001" step="0.001" value={l.quantity} onChange={e=>setLines(lines.map((x,n)=>n===i?{...x,quantity:e.target.value}:x))}/><input type="number" min="0" step="0.01" placeholder="Price" value={l.price} onChange={e=>setLines(lines.map((x,n)=>n===i?{...x,price:e.target.value}:x))}/><input className="tax" type="number" min="0" max="100" step="0.01" placeholder="Tax %" value={l.tax} onChange={e=>setLines(lines.map((x,n)=>n===i?{...x,tax:e.target.value}:x))}/>{lines.length>1&&<button type="button" className="icon" onClick={()=>setLines(lines.filter((_,n)=>n!==i))}><X size={15}/></button>}</div>)}</div><div className="form-grid"><label>Discount ({currency})<input type="number" min="0" step="0.01" value={discount} onChange={e=>setDiscount(e.target.value)}/></label><label className="wide">Notes<textarea value={notes} onChange={e=>setNotes(e.target.value)}/></label></div>{error&&<div className="error-box">{error}</div>}<div className="modal-actions"><button type="button" onClick={()=>setModal(false)}>Cancel</button><button className="primary">Create quotation</button></div></form></div>}</>
}

function Invoices({role,currency}:{role:string;currency:string}){
 const [items,setItems]=useState<Invoice[]>([]),[error,setError]=useState('')
 async function load(){try{setItems(await api<Invoice[]>('/api/v1/invoices'))}catch(e){setError(e instanceof Error?e.message:'Could not load invoices')}}useEffect(()=>{load()},[])
 async function status(x:Invoice,value:string){try{await api(`/api/v1/invoices/${x.id}/status`,{method:'PATCH',body:JSON.stringify({status:value})});await load()}catch(e){setError(e instanceof Error?e.message:'Could not update invoice')}}
 return <><header><div><p>RECEIVABLES</p><h1>Invoices</h1><span>Invoices generated from approved quotations with immutable line snapshots.</span></div></header>{error&&<div className="error-box">{error}</div>}<div className="table-wrap"><table><thead><tr><th>Invoice</th><th>Customer</th><th>Due</th><th>Total</th><th>Paid</th><th>Balance</th><th>Status</th><th>Document</th></tr></thead><tbody>{items.length===0?<Empty cols={8} icon={<ReceiptText size={28}/>} title="No invoices" text="Convert a quotation to generate the first invoice."/>:items.map(x=><tr key={x.id}><td><b>{x.number}</b></td><td>{x.customer_name}</td><td>{x.due_date||'—'}</td><td>{money(x.total_minor,currency)}</td><td>{money(x.paid_minor,currency)}</td><td><b>{money(x.balance_minor,currency)}</b></td><td>{['owner','admin','manager'].includes(role)&&!['paid','partial'].includes(x.status)?<select className="status-select" value={x.status} onChange={e=>status(x,e.target.value)}><option>draft</option><option>sent</option><option>overdue</option><option>void</option></select>:<Status value={x.status}/>}</td><td><a className="button-link" target="_blank" href={`/api/v1/documents/invoices/${x.id}/pdf`}><Download size={14}/> PDF</a></td></tr>)}</tbody></table></div></>
}

function Payments({role,currency}:{role:string;currency:string}){
 const [items,setItems]=useState<Payment[]>([]),[invoices,setInvoices]=useState<Invoice[]>([]),[modal,setModal]=useState(false),[error,setError]=useState('')
 const [invoiceID,setInvoiceID]=useState(''),[amount,setAmount]=useState(''),[method,setMethod]=useState('bank_transfer'),[reference,setReference]=useState(''),[paidAt,setPaidAt]=useState(today())
 async function load(){try{const [p,i]=await Promise.all([api<Payment[]>('/api/v1/payments'),api<Invoice[]>('/api/v1/invoices')]);setItems(p);setInvoices(i.filter(x=>x.balance_minor>0&&x.status!=='void'))}catch(e){setError(e instanceof Error?e.message:'Could not load payments')}}useEffect(()=>{load()},[])
 function pick(id:string){setInvoiceID(id);const i=invoices.find(x=>x.id===id);if(i)setAmount(String(i.balance_minor/100))}
 async function save(e:FormEvent){e.preventDefault();try{await api('/api/v1/payments',{method:'POST',body:JSON.stringify({invoice_id:invoiceID,amount_minor:toMinor(amount),method,reference,paid_at:paidAt})});setModal(false);setInvoiceID('');setAmount('');await load()}catch(e){setError(e instanceof Error?e.message:'Could not record payment')}}
 return <><header><div><p>CASH COLLECTION</p><h1>Payments</h1><span>Record collections and let invoice status update automatically.</span></div>{role!=='viewer'&&<button className="primary" onClick={()=>{setError('');setModal(true)}}><Plus size={17}/> Record payment</button>}</header>{error&&!modal&&<div className="error-box">{error}</div>}<div className="table-wrap"><table><thead><tr><th>Date</th><th>Invoice</th><th>Method</th><th>Reference</th><th>Amount</th><th>Receipt</th></tr></thead><tbody>{items.length===0?<Empty cols={6} icon={<CreditCard size={28}/>} title="No payments" text="Payments recorded against invoices will appear here."/>:items.map(x=><tr key={x.id}><td>{x.paid_at}</td><td><b>{x.invoice_number}</b></td><td className="capitalize">{x.method.replace('_',' ')}</td><td>{x.reference||'—'}</td><td><b>{money(x.amount_minor,currency)}</b></td><td><a className="button-link" target="_blank" href={`/api/v1/documents/payments/${x.id}/pdf`}><Download size={14}/> Receipt</a></td></tr>)}</tbody></table></div>{modal&&<div className="modal-backdrop" onMouseDown={()=>setModal(false)}><form className="modal" onSubmit={save} onMouseDown={e=>e.stopPropagation()}><div className="modal-head"><div><small>PAYMENT</small><h2>Record payment</h2></div><button type="button" className="icon" onClick={()=>setModal(false)}><X/></button></div><div className="form-grid"><label className="wide">Invoice<select required value={invoiceID} onChange={e=>pick(e.target.value)}><option value="">Choose invoice…</option>{invoices.map(i=><option key={i.id} value={i.id}>{i.number} — {i.customer_name} — {money(i.balance_minor,currency)}</option>)}</select></label><label>Amount ({currency})<input type="number" min="0.01" step="0.01" required value={amount} onChange={e=>setAmount(e.target.value)}/></label><label>Paid at<input type="date" required value={paidAt} onChange={e=>setPaidAt(e.target.value)}/></label><label>Method<select value={method} onChange={e=>setMethod(e.target.value)}><option value="bank_transfer">Bank transfer</option><option value="qris">QRIS</option><option value="ewallet">E-Wallet</option><option value="cash">Cash</option><option value="card">Card</option><option value="other">Other</option></select></label><label>Reference<input value={reference} onChange={e=>setReference(e.target.value)}/></label></div>{error&&<div className="error-box">{error}</div>}<div className="modal-actions"><button type="button" onClick={()=>setModal(false)}>Cancel</button><button className="primary">Save payment</button></div></form></div>}</>
}

function Status({value}:{value:string}){return <span className={`status status-${value}`}>{value}</span>}
function Empty({cols,icon,title,text}:{cols:number;icon:any;title:string;text:string}){return <tr><td colSpan={cols}><div className="empty">{icon}<b>{title}</b><span>{text}</span></div></td></tr>}

function Expenses({role,currency}:{role:string;currency:string}){
 const [items,setItems]=useState<Expense[]>([]),[modal,setModal]=useState(false),[error,setError]=useState('')
 const [category,setCategory]=useState('Operations'),[date,setDate]=useState(today()),[vendor,setVendor]=useState(''),[description,setDescription]=useState(''),[amount,setAmount]=useState(''),[method,setMethod]=useState('bank_transfer'),[reference,setReference]=useState('')
 async function load(){try{setItems(await api<Expense[]>('/api/v1/expenses'))}catch(e){setError(e instanceof Error?e.message:'Could not load expenses')}}useEffect(()=>{load()},[])
 async function save(e:FormEvent){e.preventDefault();try{await api('/api/v1/expenses',{method:'POST',body:JSON.stringify({category,expense_date:date,vendor,description,amount_minor:toMinor(amount),payment_method:method,reference})});setModal(false);setVendor('');setDescription('');setAmount('');setReference('');await load()}catch(e){setError(e instanceof Error?e.message:'Could not save expense')}}
 async function remove(x:Expense){if(!confirm(`Delete expense “${x.description}”?`))return;try{await api(`/api/v1/expenses/${x.id}`,{method:'DELETE'});await load()}catch(e){setError(e instanceof Error?e.message:'Could not delete expense')}}
 return <><header><div><p>OPERATING COSTS</p><h1>Expenses</h1><span>Track business spending using the same integer-money model as invoices and payments.</span></div>{role!=='viewer'&&<button className="primary" onClick={()=>{setError('');setModal(true)}}><Plus size={17}/> New expense</button>}</header>{error&&!modal&&<div className="error-box">{error}</div>}<div className="table-wrap"><table><thead><tr><th>Date</th><th>Category</th><th>Description</th><th>Vendor</th><th>Method</th><th>Amount</th><th></th></tr></thead><tbody>{items.length===0?<Empty cols={7} icon={<WalletCards size={28}/>} title="No expenses" text="Record operational spending to unlock financial reporting."/>:items.map(x=><tr key={x.id}><td>{x.expense_date}</td><td>{x.category||'Uncategorized'}</td><td><b>{x.description}</b><small>{x.reference||''}</small></td><td>{x.vendor||'—'}</td><td className="capitalize">{x.payment_method.replace('_',' ')}</td><td><b>{money(x.amount_minor,currency)}</b></td><td>{['owner','admin','manager'].includes(role)&&<button className="danger" onClick={()=>remove(x)}><Trash2 size={15}/></button>}</td></tr>)}</tbody></table></div>{modal&&<div className="modal-backdrop" onMouseDown={()=>setModal(false)}><form className="modal" onSubmit={save} onMouseDown={e=>e.stopPropagation()}><div className="modal-head"><div><small>EXPENSE</small><h2>Record expense</h2></div><button type="button" className="icon" onClick={()=>setModal(false)}><X/></button></div><div className="form-grid"><label>Date<input type="date" required value={date} onChange={e=>setDate(e.target.value)}/></label><label>Category<input value={category} onChange={e=>setCategory(e.target.value)} placeholder="Operations"/></label><label className="wide">Description<input required value={description} onChange={e=>setDescription(e.target.value)} placeholder="Office internet"/></label><label>Vendor<input value={vendor} onChange={e=>setVendor(e.target.value)}/></label><label>Amount ({currency})<input type="number" min="0.01" step="0.01" required value={amount} onChange={e=>setAmount(e.target.value)}/></label><label>Method<select value={method} onChange={e=>setMethod(e.target.value)}><option value="bank_transfer">Bank transfer</option><option value="qris">QRIS</option><option value="ewallet">E-Wallet</option><option value="cash">Cash</option><option value="card">Card</option><option value="other">Other</option></select></label><label>Reference<input value={reference} onChange={e=>setReference(e.target.value)}/></label></div>{error&&<div className="error-box">{error}</div>}<div className="modal-actions"><button type="button" onClick={()=>setModal(false)}>Cancel</button><button className="primary">Save expense</button></div></form></div>}</>
}

function Reports({currency}:{currency:string}){const[d,setD]=useState<FinanceReport|null>(null),[error,setError]=useState('');useEffect(()=>{api<FinanceReport>('/api/v1/reports/finance').then(setD).catch(e=>setError(e.message))},[]);if(!d)return <>{error?<div className="error-box">{error}</div>:<Splash/>}</>;return <><header><div><p>FINANCIAL REPORT</p><h1>Business performance</h1><span>A cash-oriented operating snapshot from invoices, collections and recorded expenses.</span></div></header><section className="stats"><Stat label="Invoiced" value={money(d.invoiced_minor,currency)}/><Stat label="Collected" value={money(d.collected_minor,currency)}/><Stat label="Expenses" value={money(d.expense_minor,currency)}/><Stat label="Cash result" value={money(d.cash_result_minor,currency)}/></section><section className="grid"><article className="panel"><small>ACCOUNTS RECEIVABLE</small><h2>{money(d.receivable_minor,currency)}</h2><p className="muted">{d.open_invoices} invoice(s) remain open. This is an operational report, not a full accrual accounting statement.</p></article><article className="panel"><small>FORMULA</small><h2>Collected − Expenses</h2><p className="muted">Cash result intentionally stays simple for small-business visibility. Full ledger/accounting can be introduced as a separate module later.</p></article></section></>}

function SettingsPage({me}:{me:Me}){const empty:BusinessProfile={legal_name:'',tax_id:'',email:'',phone:'',address:'',bank_name:'',bank_account_name:'',bank_account_number:'',qris_text:''};const[form,setForm]=useState<BusinessProfile>(empty),[saved,setSaved]=useState(false),[error,setError]=useState('');useEffect(()=>{api<BusinessProfile>('/api/v1/company/profile').then(setForm).catch(e=>setError(e.message))},[]);const canEdit=['owner','admin'].includes(me.company!.role);async function save(e:FormEvent){e.preventDefault();setError('');try{await api('/api/v1/company/profile',{method:'PUT',body:JSON.stringify(form)});setSaved(true);setTimeout(()=>setSaved(false),1800)}catch(e){setError(e instanceof Error?e.message:'Could not save profile')}}return <><header><div><p>BUSINESS PROFILE</p><h1>{me.company!.name}</h1><span>Legal identity and payment details used on generated documents.</span></div></header><form className="panel profile-form" onSubmit={save}><div className="form-grid"><label>Legal name<input disabled={!canEdit} value={form.legal_name} onChange={e=>setForm({...form,legal_name:e.target.value})}/></label><label>Tax ID / NPWP<input disabled={!canEdit} value={form.tax_id} onChange={e=>setForm({...form,tax_id:e.target.value})}/></label><label>Email<input disabled={!canEdit} value={form.email} onChange={e=>setForm({...form,email:e.target.value})}/></label><label>Phone<input disabled={!canEdit} value={form.phone} onChange={e=>setForm({...form,phone:e.target.value})}/></label><label className="wide">Address<textarea disabled={!canEdit} value={form.address} onChange={e=>setForm({...form,address:e.target.value})}/></label><label>Bank<input disabled={!canEdit} value={form.bank_name} onChange={e=>setForm({...form,bank_name:e.target.value})}/></label><label>Account number<input disabled={!canEdit} value={form.bank_account_number} onChange={e=>setForm({...form,bank_account_number:e.target.value})}/></label><label className="wide">Account name<input disabled={!canEdit} value={form.bank_account_name} onChange={e=>setForm({...form,bank_account_name:e.target.value})}/></label><label className="wide">QRIS / payment instructions<textarea disabled={!canEdit} value={form.qris_text} onChange={e=>setForm({...form,qris_text:e.target.value})}/></label></div>{error&&<div className="error-box">{error}</div>}{saved&&<div className="success-box">Business profile saved.</div>}{canEdit&&<div className="modal-actions"><button className="primary">Save business profile</button></div>}</form></>}

function Stat({
  label,
  value,
  detail,
}: {
  label: string
  value: string
  detail?: string
}) {
  return (
    <article>
      <small>{label}</small>
      <strong>{value}</strong>
      {detail && <span>{detail}</span>}
    </article>
  )
}