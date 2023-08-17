import React, { useEffect, useState } from 'react';
import DatePicker from 'react-datepicker'
import Modal from 'react-modal';
import DateTimePicker from './DateTimePicker';

export const PAGE_SIZE = 50;
const FlagTableComponent = ({ data }) => {
  const [flags, setFlags] = useState(data)
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [filteredFlags, setFilteredFlags] = useState([...flags]);
  const [filterUsername, setFilterUsername] = useState('');
  const [filterFlag, setFilterFlag] = useState('');
  const [filterStatus, setFilterStatus] = useState("All");
  const [filterTeam, setFilterTeam] = useState("All")
  const [filterExploitName, setFilterExploitName] = useState("All")
  const [filterCheckSystemResponse, setFilterCheckSystemResponse] = useState("")
  const [fromDateTime, setFromDateTime] = useState("")
  const [toDateTime, setToDateTime] = useState("")

  /*useEffect(() => {
    // Update the chart data whenever the "data" prop changes
    setFlags(data);
    handleFilter();
  }, [data]);*/

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

    var filtered = flags
    if(filterUsername != '')
      filtered = filtered.filter(flag => flag.username === filterUsername);
    if(filterFlag != ''){
      filtered = filtered.filter(flag => flag.flag.includes(filterFlag))
    }
    if (filterStatus != "All"){
      filtered = filtered.filter(flag => flag.status == filterStatus)
    }
    if (filterTeam != "All"){
      filtered = filtered.filter(flag => flag.team_ip == filterTeam)
    }
    if (filterExploitName != "All"){
      filtered = filtered.filter(flag => flag.exploit_name == filterExploitName)
    }
    if (filterCheckSystemResponse != ""){
      filtered = filtered.filter(flag => flag.server_response == filterCheckSystemResponse)
    }
    if (fromDateTime != ""){
      filtered = filtered.filter(flag => flag.time >= fromDateTime )
    }
    if(toDateTime != ""){
      filtered = filtered.filter(flag => flag.time <= toDateTime)
    }

    
    setFilteredFlags(filtered);
    setCurrentPage(1)
    closeModal(); // Close the modal after filtering
  };
  // Calculate the start and end index of the current page
  const startIndex = (currentPage - 1) * PAGE_SIZE;
  const endIndex = startIndex + PAGE_SIZE;

  const visibleFlags = filteredFlags.slice(startIndex, endIndex);

  const totalPages = Math.ceil(filteredFlags.length / PAGE_SIZE);

  return (
    <div>
      <h1>Flag Table</h1>
      <button onClick={openModal}>Filter Flags</button>

      <table>
        <thead>
          <tr>
            <th>Flag</th>
            <th>Status</th>
            <th>Server Response</th>
            <th>Exploit Name</th>
            <th>Time</th>
            <th>Team ip</th>
            <th>Username</th>
          </tr>
        </thead>
        <tbody>
          {visibleFlags.map(flag => (
            <tr key={flag.flag}>
              <td>{flag.flag}</td>
              <td>{flag.status}</td>
              <td>{flag.server_response}</td>
              <td>{flag.exploit_name}</td>
              <td>{flag.time}</td>
              <td>{flag.team_ip}</td>
              <td>{flag.username}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div>
        {Array.from({ length: totalPages }, (_, index) => (
          <button
            key={index + 1}
            onClick={() => setCurrentPage(index + 1)}
            style={{ fontWeight: currentPage === index + 1 ? 'bold' : 'normal' }}
          >
            {index + 1}
          </button>
        ))}
      </div>

      <Modal isOpen={isModalOpen} onRequestClose={closeModal} >
        <div className='modal_content'>
          <label for="exploitSelect">Select an exploit:</label>
          <select id="exploitSelect" onChange={(e=>setFilterExploitName(e.target.value))}>
              <option value="All">All</option>
              {visibleFlags.map((flag)=>flag.exploit_name).filter((value, index, self) => self.indexOf(value) === index).map((exploit_name)=>(
                <option value={exploit_name}>{exploit_name}</option>
            ))}
          </select>
          <label for="teamSelect">Select a team:</label>
          <select id="teamSelect" onChange={(e=>setFilterTeam(e.target.value))}>
              <option value="All">All</option>
              {visibleFlags.map((flag)=>flag.team_ip).filter((value, index, self) => self.indexOf(value) === index).map((team_ip)=>(
                <option value={team_ip}>{team_ip}</option>
            ))}
          </select>
          <label for="responseSelect">Select a response:</label>
          <select id="responseSelect" onChange={(e=>setFilterCheckSystemResponse(e.target.value))}>
              <option value=""></option>
              {visibleFlags.map((flag)=>flag.server_response).filter((value, index, self) => self.indexOf(value) === index).map((server_response)=>(
                <option value={server_response}>{server_response}</option>
            ))}
          </select>
          <label for="statusSelect">Select a status:</label>
          <select id="statusSelect" onChange={(e=>setFilterStatus(e.target.value))}>
              <option value="All">All</option>
              {visibleFlags.map((flag)=>flag.status).filter((value, index, self) => self.indexOf(value) === index).map((status)=>(
                <option value={status}>{status}</option>
            ))}
          </select>
          <label for="fromDateTime">Select flags from:</label>
          <div className="date-time-picker-component">
            <DateTimePicker selectedDateTime={fromDateTime} setSelectedDateTime={setFromDateTime}/>
          </div>
      
          <label for="fromDateTime">Select flags to:</label>
          <div className="date-time-picker-component">
            <DateTimePicker selectedDateTime={toDateTime} setSelectedDateTime={setToDateTime}/>
          </div>
      
          <label for="flag_filter">Enter the flag you want to search for:</label>
          <input
              id="flag_filter"
              type="text"
              placeholder="Flag sustring"
              value={filterFlag}
              onChange={e => setFilterFlag(e.target.value)}
          />
          <label for="username_filter">Enter the username you want to filter:</label>
          <input
            id="username_filter"
            type="text"
            placeholder="Filter by username"
            value={filterUsername}
            onChange={e => setFilterUsername(e.target.value)}
          />
          <button onClick={handleFilter}>Apply Filter</button>
          <button onClick={closeModal}>Close</button>
        </div>
      </Modal>
    </div>
  );
};

export default FlagTableComponent;
