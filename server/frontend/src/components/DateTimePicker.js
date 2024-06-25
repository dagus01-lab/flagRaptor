import React, { useEffect, useState } from 'react';
import Calendar from 'react-calendar';
import 'react-calendar/dist/Calendar.css';
import dayjs, { Dayjs } from "dayjs";
import { DemoContainer } from "@mui/x-date-pickers/internals/demo";
import { LocalizationProvider } from "@mui/x-date-pickers/LocalizationProvider";
import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";
import { TimePicker } from "@mui/x-date-pickers/TimePicker";

const DateTimePicker = ({selectedDateTime, setSelectedDateTime}) => {
  const [selectedDate, setSelectedDate] = useState(dayjs(new Date()));
  const [selectedTime, setSelectedTime] = useState(dayjs(new Date()));
  const [showCalendar, setShowCalendar] = useState(false);
  const [showClock, setShowClock] = useState(false);

  const formatDateToCustomString = (date, time) => {
    return date.format('YYYY-MM-DD')+' '+time.format('HH:mm:ss');
  }

  useEffect(() =>{
    setSelectedDateTime(formatDateToCustomString(selectedDate, selectedTime))
  }, [selectedDate, selectedTime])

  const handleInputClick = () => {
    setShowCalendar(true);
  };

  const handleCalendarChange = date => {
    setSelectedDate(dayjs(date));
    setShowClock(true);
    setShowCalendar(false);
  };


  return (
    <div className="date-time-picker-container">
      <input
        type="text"
        className="date-time-input"
        onClick={handleInputClick}
        value={selectedDateTime}
        readOnly
      />
      {showCalendar && (
        <div className="date-picker">
          <h4>Select Date</h4>
          <Calendar
            onChange={handleCalendarChange}
            value={selectedDate.toDate()}
          />
        </div>
      )}
      {showClock && (
        <div className="time-picker" >
          <h4>Select Time</h4>
          <LocalizationProvider dateAdapter={AdapterDayjs}>
            <DemoContainer components={["TimePicker", "TimePicker"]}>
                <TimePicker
                value={selectedTime}
                onChange={(newValue) => setSelectedTime(newValue)}
                />
            </DemoContainer>
         </LocalizationProvider>
         <button onClick={()=>{setShowClock(false)}}>Confirm</button>
        </div>
        )
      }
    </div>
  );
};

export default DateTimePicker;
