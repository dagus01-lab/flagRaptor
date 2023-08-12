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

export const options = {
  responsive: true,
  plugins: {
    legend: {
      position: 'top',
    },
    title: {
      display: true,
      text: 'Your flags',
    },
  },
};

const BarChartComponent = ({label,data}) => {
  const [chartData, setChartData] = useState(data);

  useEffect(() => {
    // Update the chart data whenever the "data" prop changes
    setChartData(data);
  }, [data]);

  return (
    <div>
      <h4>{label}</h4>
      <Bar options={options} data={data}/>
    </div>
  );
};

export default BarChartComponent;
