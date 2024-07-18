package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/xuri/excelize/v2"
	"xyz.com/practiseGO/Concurrency"
	"xyz.com/practiseGO/RedisClient"
)

var XLSXfilepath string = "employees.xlsx"

func mergeArray(array []float64, l int, mid int, r int) {
	ans := make([]float64, r-l+1)
	p := 0
	s1 := l
	s2 := mid + 1
	for s1 <= mid && s2 <= r {
		if array[s2] <= array[s1] {
			ans[p] = array[s2]
			s2++
		} else {
			ans[p] = array[s1]
			s1++
		}
		p++
	}
	for s1 <= mid {
		ans[p] = array[s1]
		s1++
		p++
	}
	for s2 <= r {
		ans[p] = array[s2]
		s2++
		p++
	}
	copy(array[l:r+1], ans)
}
func mergeSort(array []float64, l int, r int) {
	if l >= r {
		return
	}
	var mid int = l + (r-l)/2
	mergeSort(array, l, mid)
	mergeSort(array, mid+1, r)
	mergeArray(array, l, mid, r)
}

func binarySearch(array []float64, val float64) int {
	l := 0
	r := len(array) - 1
	for l <= r {
		var mid int = l + (r-l)/2
		if array[mid] == val {
			return mid
		} else if val > array[mid] {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	return -1
}

type Employee struct {
	ID         int    `json:"employeeid"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Position   string `json:"position"`
	Salary     int    `json:"salary"`
}

func openEmployeeXlFile() (*excelize.File, error) {
	filePath := XLSXfilepath
	xlsx, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening Excel file: %v", err)
	}
	return xlsx, nil
}
func getEmployeeByIdXl(searchID string) (map[string]interface{}, error) {
	xlsx, err := openEmployeeXlFile()
	if err != nil {
		return make(map[string]interface{}, 0), fmt.Errorf("%v", err)
	}
	rows, err := xlsx.GetRows("Employees")
	if err != nil {
		return make(map[string]interface{}, 0), fmt.Errorf("%v", err)
	}
	var employeeIds []float64

	for _, row := range rows {
		empId, err := strconv.Atoi(row[0])
		if err != nil {
			fmt.Println("error in converting")
		}
		employeeIds = append(employeeIds, float64(empId))
	}
	searchIDInt, err2 := strconv.Atoi(searchID)
	if err2 != nil {
		fmt.Println("error in converting")
	}

	idIndex := binarySearch(employeeIds, float64(searchIDInt))
	if idIndex == -1 {
		return make(map[string]interface{}), fmt.Errorf("no employee id found")
	}
	empData := map[string]interface{}{
		"employeeid": rows[idIndex][0],
		"name":       rows[idIndex][1],
		"department": rows[idIndex][2],
		"position":   rows[idIndex][3],
		"salary":     rows[idIndex][4],
	}
	return empData, nil
}

func getEmployeeById(c *gin.Context) {
	searchID := c.Query("employeeId")
	type1 := c.Query("db")
	if type1 == "" || type1 == "redis" {
		redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
		rclient := RedisClient.NewRedisClient(&redisAddr, nil, nil)
		employeeString, err := rclient.Get(context.Background(), searchID).Result()
		if err == redis.Nil {
			empData, err3 := getEmployeeByIdXl(searchID)
			if err3 != nil {
				c.JSON(200, gin.H{"error": "No Employee found"})
				return
			}
			jsonData, err := json.Marshal(empData)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to generate JSON"})
				return
			}
			err = rclient.Set(context.Background(), searchID, jsonData, 0).Err()
			if err != nil {
				fmt.Printf("Error while setting in redis : %v", err)
			}
			c.Header("Content-Type", "application/json")
			c.String(200, string(jsonData))
		} else if err != nil {
			fmt.Printf("Error with redis : %v", err)
		} else {
			c.String(200, employeeString)
			return
		}

	} else {
		empData, err3 := getEmployeeByIdXl(searchID)
		if err3 != nil {
			c.JSON(200, gin.H{"error": "No Employee found"})
			return
		}
		jsonData, err := json.Marshal(empData)
		if err != nil {
			c.JSON(200, gin.H{"error": "Failed to generate JSON"})
			return
		}
		c.Header("Content-Type", "application/json")
		c.String(200, string(jsonData))
	}
}
func addEmployeeToXl(c *gin.Context) {
	var newEmployee Employee
	if err := c.ShouldBindJSON(&newEmployee); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Open or create the Excel file
	filename, err := openEmployeeXlFile()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	sheetName := "Employees"
	// Find the next available row
	rows, err := filename.GetRows(sheetName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read the Excel file"})
		return
	}
	nextRow := len(rows) + 1

	// Write the new employee data to the next row
	filename.SetCellValue(sheetName, "A"+strconv.Itoa(nextRow), newEmployee.ID)
	filename.SetCellValue(sheetName, "B"+strconv.Itoa(nextRow), newEmployee.Name)
	filename.SetCellValue(sheetName, "C"+strconv.Itoa(nextRow), newEmployee.Department)
	filename.SetCellValue(sheetName, "D"+strconv.Itoa(nextRow), newEmployee.Position)
	filename.SetCellValue(sheetName, "E"+strconv.Itoa(nextRow), newEmployee.Salary)

	// Save the Excel file
	if err := filename.SaveAs(XLSXfilepath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save the Excel file"})
		return
	}

	// Respond with success
	c.JSON(200, gin.H{"message": "Employee added successfully"})
}

func FindSumUsingNWorkersReqHandler(c *gin.Context) {

	num, err := strconv.Atoi(c.Param("num"))
	if err != nil {
		c.JSON(500, gin.H{"message": "Couldnt extract num in findsum"})
	}
	noOfWorkers, err := strconv.Atoi(c.Param("noOfWorkers"))
	if err != nil {
		c.JSON(500, gin.H{"message": "Couldnt extract num in findsum"})
	}
	val, timeTaken := Concurrency.FindSumUsingNWorkers(num, noOfWorkers)
	c.JSON(200, gin.H{"value": val, "timeTaken": timeTaken})
}

func deleteEmployeeById(c *gin.Context) {
	id := c.Param("id")
	filename, err := openEmployeeXlFile()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	sheetName := "Employees"
	rows, err := filename.GetRows(sheetName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read the Excel file: " + err.Error()})
		return
	}
	var rowIndexToDelete int
	idFound := false
	for rowIndex, row := range rows {
		if rowIndex == 0 {
			continue
		}
		if len(row) > 0 && row[0] == id {
			rowIndexToDelete = rowIndex + 1
			idFound = true
			break
		}
	}

	if !idFound {
		c.JSON(404, gin.H{"error": "Employee ID not found"})
		return
	}

	filename.RemoveRow(sheetName, rowIndexToDelete)
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	rclient := RedisClient.NewRedisClient(&redisAddr, nil, nil)
	isExists, err := rclient.Exists(context.Background(), id).Result()
	if isExists > 0 {
		if rclient.Del(context.Background(), id).Err() != nil {
			c.JSON(500, gin.H{"error": "Error deleting from redis " + err.Error()})

		}
	}
	if err := filename.SaveAs(XLSXfilepath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save the Excel file: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Employee deleted successfully"})
}

func main() {
	// array := []float64{5.0, 11, 90, 87, 65, 100, 1001}
	// mergeSort(array, 0, len(array)-1)
	fmt.Println("hi")
	r := gin.Default()
	r.GET("/getEmployeeById", getEmployeeById)
	r.POST("/addEmployee", addEmployeeToXl)
	r.GET("/sumUsingNWorkers/:num/:noOfWorkers", FindSumUsingNWorkersReqHandler)
	//r.PosT("")
	r.DELETE("/deleteEmployeeById/:id", deleteEmployeeById)
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
