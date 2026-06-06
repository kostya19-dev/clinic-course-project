export const demoClinicContext = {
  adminUserId: "00000000-0000-0000-0000-000000000101",
  patientUserId: "00000000-0000-0000-0000-000000000102",
  doctorUserId: "00000000-0000-0000-0000-000000000103",
  doctorId: "00000000-0000-0000-0000-000000000301",
  doctorName: "Иван Иванов",
  cabinetNumber: "101",
  defaultScheduleFromDate: "2030-01-15",
  defaultScheduleToDate: "2030-01-16",
  defaultAppointmentAt: "2030-01-15T10:00",
  patients: [
    {
      id: "00000000-0000-0000-0000-000000000601",
      fullName: "Анна Сергеевна Петрова",
    },
    {
      id: "00000000-0000-0000-0000-000000000602",
      fullName: "Павел Олегович Сидоров",
    },
  ],
  services: [
    {
      id: "00000000-0000-0000-0000-000000000401",
      code: "CONSULT_CARDIO",
      name: "Консультация кардиолога",
      price: 1500,
    },
    {
      id: "00000000-0000-0000-0000-000000000402",
      code: "ECG_STANDARD",
      name: "Стандартная ЭКГ",
      price: 900,
    },
  ],
};
