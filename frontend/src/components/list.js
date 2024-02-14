import '../App.css'

export const List = ({listData, handleImageClick})=>{
    return(
        <div className='list'>
                {
                    listData?.map((image, index) =>{
                        return(
                            <div className="list-container" key={`${image.url}-${index}`}>
                                <div>{image.description}</div>
                                <button onClick={()=>{handleImageClick(image.url)}}>{`View`}</button>
                            </div>
                        )
                    })
                }
        </div>
    )
}