import { useState } from "react"

export const Search = ()=>{

    const [searchVal, setSearchVal] = useState('')

    const handleChange = (event)=>{
        const val = event.target.value
        setSearchVal(val)
    }

    return(
        <div className="search-container">
            <label>{`Search`}</label>
            <input type="text" value={searchVal} onChange={handleChange}/>
        </div>
    )
}