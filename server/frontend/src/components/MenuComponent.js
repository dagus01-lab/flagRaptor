import React from "react"
const MenuComponent = ({setStatus, setLoginFunction}) => {
    return (
        <div className="menu-bar">
          <div className="menu-item" onClick={()=>setStatus("home")}>Home</div>
          <div className="menu-item" onClick={()=>setStatus("explore")}>Explore</div>
          <div className="menu-item" onClick={()=>setStatus("submit")}>Submit</div>
          <div className="menu-item" onClick={()=>setStatus("logout")}>Logout</div>
        </div>
    )
};
export default MenuComponent