import React, {useEffect, useState} from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

export const optionsTop = {
  responsive: true,
  plugins: {
    legend: {
      position: 'top',
    },
  },
};
export const optionsCenter = {
  responsive: true,
  plugins: {
    legend: {
      position: 'center',
    },
  },
};

const BarChartComponent = ({label,data,options}) => {
  const [chartData, setChartData] = useState(data);

  useEffect(() => {
    // Update the chart data whenever the "data" prop changes
    setChartData(data);
  }, [data]);

  return (
    <div>
      <h4>{label}</h4>
      <Bar options={options} data={chartData}/>
    </div>
  );
};

export default BarChartComponent;
