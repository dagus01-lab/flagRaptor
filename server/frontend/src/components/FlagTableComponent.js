import React, { useEffect, useState } from 'react';
import Modal from 'react-modal';
import DateTimePicker from './DateTimePicker';
import { Input } from '@mui/base/Input'
import { Button } from '@mui/material'
import { DataGrid, GridColDef, GridValueGetterParams } from '@mui/x-data-grid'

const FlagTableComponent = ({ data }) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [filteredFlags, setFilteredFlags] = useState([]);
  const [columns] = useState([
    {field: 'flag', headerName: 'Flag', width: 350},
    {field: 'status', headerName: 'Status', width: 100},
    {field: 'server_response', headerName: 'Server Response', width: 100},
    {field: 'exploit_name', headerName: 'Exploit Name', width: 150},
    {field: 'time', headerName: 'Time', width: 200},
    {field: 'team_ip', headerName: 'Team IP', width:100},
    {field: 'username', headerName: 'Username', width:100},
  ])
  const [fromDateTime, setFromDateTime] = useState("")
  const [toDateTime, setToDateTime] = useState("")

  useEffect(() => {
    handleFilter()
  }, []);

  const openModal = () => {
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
  };

  const handleFilter = () => {
    //check if fromDateTime<toDateTime
    if(fromDateTime > toDateTime){
      alert("The first datetime must be before the second one")
      return
    }
    console.log("data: "+data)

    var filtered = data
    if (fromDateTime != ""){
      filtered = filtered.filter(flag => flag.time >= fromDateTime )
    }
    if(toDateTime != ""){
      filtered = filtered.filter(flag => flag.time <= toDateTime)
    }
    
    setFilteredFlags(filtered.map(flag => {
      var f = {...flag}
      f.id = data.indexOf(flag)
      return f
    }))
    closeModal(); // Close the modal after filtering
  };

  return (
    
    <div>
      <h1>Flag Table</h1>
      <Button onClick={openModal}>Filter Flags</Button>

      <DataGrid
        rows={filteredFlags}
        columns={columns}
        initialState={{
          pagination: {
            paginationModel: { page: 0, pageSize: 50 },
          },
        }}
        pageSizeOptions={[50, 100]}
      />

      <Modal isOpen={isModalOpen} onRequestClose={closeModal} >
        <div className='modal_content'>
          <label for="fromDateTime">Select flags from:</label>
          <div className="date-time-picker-component">
            <DateTimePicker selectedDateTime={fromDateTime} setSelectedDateTime={setFromDateTime}/>
          </div>
      
          <label for="fromDateTime">Select flags to:</label>
          <div className="date-time-picker-component">
            <DateTimePicker selectedDateTime={toDateTime} setSelectedDateTime={setToDateTime}/>
          </div>
          <Button onClick={handleFilter}>Apply Filter</Button>
          <Button onClick={closeModal}>Close</Button>
        </div>
      </Modal>
    </div>
  );
};

export default FlagTableComponent;
