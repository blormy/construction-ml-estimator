package main

import "time"

type Equipment struct {
	Type     string `json:"type"`
	Tag      string `json:"tag"`
	Quantity int    `json:"quantity"`

	// насос
	FlowRate    *float64 `json:"flow_rate,omitempty"`
	FluidHead   *float64 `json:"fluid_head,omitempty"`
	RPM         *float64 `json:"rpm,omitempty"`
	SpecGravity *float64 `json:"spec_gravity,omitempty"`
	PowerKW     *float64 `json:"power_kw,omitempty"`

	// конвейер
	ConveyorLength   *float64 `json:"conveyor_length,omitempty"`
	BeltWidth        *float64 `json:"belt_width,omitempty"`
	ConveyorFlowRate *float64 `json:"conveyor_flow_rate,omitempty"`

	// аппараты, емкости и колонны
	VesselDiameter               *float64 `json:"vessel_diameter,omitempty"`
	DesignTangentToTangentLength *float64 `json:"design_tangent_to_tangent_length,omitempty"`
	VesselTangentToTangentHeight *float64 `json:"vessel_tangent_to_tangent_height,omitempty"`
	DesignGaugePressure          *float64 `json:"design_gauge_pressure,omitempty"`
	DesignTemperature            *float64 `json:"design_temperature,omitempty"`
	SkirtHeight                  *float64 `json:"skirt_height,omitempty"`
	VesselLegHeight              *float64 `json:"vessel_leg_height,omitempty"`
	NumberOfTrays                *float64 `json:"number_of_trays,omitempty"`

	// теплообменники (u-tube)
	ShellDiameter   *float64 `json:"shell_diameter,omitempty"`
	TubeOutDiameter *float64 `json:"tube_out_diameter,omitempty"`
	TubeLen         *float64 `json:"tube_len,omitempty"`
	TubeDesPres     *float64 `json:"tube_des_pres,omitempty"`
	HeatArea        *float64 `json:"heat_area,omitempty"`

	// коробчатая технологическая печь (box furnace)
	Duty                *float64 `json:"duty,omitempty"`
	StandardGasFlowRate *float64 `json:"standard_gas_flow_rate,omitempty"`

	// центробежный компрессор (centrifugal compressor)
	ActualGasFlowRateInlet    *float64 `json:"actual_gas_flow_rate_inlet,omitempty"`
	DesignGaugePressureInlet  *float64 `json:"design_gauge_pressure_inlet,omitempty"`
	DesignGaugePressureOutlet *float64 `json:"design_gauge_pressure_outlet,omitempty"`
	DriverPower               *float64 `json:"driver_power,omitempty"`

	CalculatedWeight float64 `json:"calculated_weight"`
}

type Project struct {
	Name      string      `json:"name"`
	Equipment []Equipment `json:"equipment"`
	// поля для облачной синхронизации
	CloudID   uint      `json:"cloud_id,omitempty"`
	TeamID    *uint     `json:"team_id,omitempty"`
	OwnerID   uint      `json:"owner_id,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// кэшированные данные из облака для отображения в списке
	CloudItemCount   int     `json:"-"`
	CloudTotalWeight float64 `json:"-"`
}

func (p Project) TotalWeight() float64 {
	if len(p.Equipment) == 0 && p.CloudTotalWeight > 0 {
		return p.CloudTotalWeight
	}
	var total float64
	for _, eq := range p.Equipment {
		total += eq.CalculatedWeight * float64(eq.Quantity)
	}
	return total
}

func (p Project) EquipmentCount() int {
	if len(p.Equipment) == 0 && p.CloudItemCount > 0 {
		return p.CloudItemCount
	}
	var count int
	for _, eq := range p.Equipment {
		count += eq.Quantity
	}
	return count
}

type AppData struct {
	Projects []Project `json:"projects"`
}

// облачные dto для команд

// команда из облачного api
type CloudTeam struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	OwnerID uint   `json:"owner_id"`
}

// участник команды из облачного api
type CloudTeamMember struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// проект из облачного api для команд
type CloudProject struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	TeamID  *uint  `json:"team_id,omitempty"`
	OwnerID uint   `json:"owner_id"`
}
