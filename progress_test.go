package progress_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"moul.io/progress"
	"moul.io/u"
)

func Example() {
	// initialize a new progress.Progress
	prog := progress.New()
	prog.AddStep("init").SetDescription("initialize")
	prog.AddStep("step1").SetDescription("step 1")
	prog.AddStep("step2").SetData([]string{"hello", "world"}).SetDescription("step 2")
	prog.AddStep("step3")
	prog.AddStep("finish")

	// automatically mark the last step as done when the function quit
	defer prog.Get("finish").Done()

	// mark init as Done
	prog.Get("init").Done()

	// mark step1 as started
	prog.Get("step1").SetData(42).Start()

	// then, mark it as done + attach custom data
	prog.Get("step1").SetData(1337).Done()

	// mark step2 as started
	prog.Get("step2").Start()

	fmt.Println(u.PrettyJSON(prog))

	// outputs something like this:
	// {
	//  "steps": [
	//    {
	//      "id": "init",
	//      "description": "initialize",
	//      "started_at": "2020-12-22T20:26:05.717427484+01:00",
	//      "done_at": "2020-12-22T20:26:05.717427484+01:00",
	//      "state": "done"
	//    },
	//    {
	//      "id": "step1",
	//      "description": "step 1",
	//      "started_at": "2020-12-22T20:26:05.71742797+01:00",
	//      "done_at": "2020-12-22T20:26:05.717428258+01:00",
	//      "state": "done",
	//      "data": 1337,
	//      "duration": 286
	//    },
	//    {
	//      "id": "step2",
	//      "description": "step 2",
	//      "started_at": "2020-12-22T20:26:05.71742865+01:00",
	//      "state": "in progress",
	//      "data": [
	//        "hello",
	//        "world"
	//      ],
	//      "duration": 496251
	//    },
	//    {
	//      "id": "step3"
	//    },
	//    {
	//      "id": "finish"
	//    }
	//  ],
	//  "created_at": "2020-12-22T20:26:05.717423018+01:00",
	//  "snapshot": {
	//    "state": "in progress",
	//    "doing": "step 2",
	//    "not_started": 2,
	//    "in_progress": 1,
	//    "completed": 2,
	//    "total": 5,
	//    "percent": 50,
	//    "total_duration": 25935,
	//    "started_at": "2020-12-22T20:26:05.717427484+01:00"
	//  }
	//}
}

func TestFlow(t *testing.T) {
	// initialize a new progress
	prog := progress.New()
	{
		require.NotEmpty(t, prog)
		require.Empty(t, prog.Steps)
		require.NotZero(t, prog.CreatedAt)
		require.True(t, prog.CreatedAt.Before(time.Now()))
		snapshot := prog.Snapshot()
		require.Equal(t, progress.StateNotStarted, snapshot.State)
		require.Equal(t, 0, snapshot.Total)
		require.Equal(t, 0, snapshot.Completed)
		require.Equal(t, 0, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.Equal(t, float64(0), snapshot.Percent)
		require.Nil(t, prog.Get("step1"))
	}

	// add a first step
	{
		prog.AddStep("step1")
		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 1)
		require.True(t, prog.CreatedAt.Before(time.Now()))

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateNotStarted, snapshot.State)
		require.Equal(t, 1, snapshot.Total)
		require.Equal(t, 0, snapshot.Completed)
		require.Equal(t, 1, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.NotNil(t, prog.Get("step1"))
		require.Equal(t, float64(0), snapshot.Percent)

		step1 := prog.Get("step1")
		require.NotNil(t, step1)
		require.Equal(t, step1.State, progress.StateNotStarted)
		require.Empty(t, step1.Description)
		step1.SetDescription("hello")
		require.Equal(t, "hello", step1.Description)
	}

	// add a second step
	{
		prog.AddStep("step2")
		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 2)

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateNotStarted, snapshot.State)
		require.Equal(t, 2, snapshot.Total)
		require.Equal(t, 0, snapshot.Completed)
		require.Equal(t, 2, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.NotNil(t, prog.Get("step2"))
		require.Equal(t, float64(0), snapshot.Percent)
	}

	// start the first step
	{
		step1 := prog.Get("step1")
		step1.Start()

		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 2)

		snapshot := prog.Snapshot()
		require.Equal(t, "hello", snapshot.Doing)
		require.Equal(t, progress.StateInProgress, snapshot.State)
		require.Equal(t, 2, snapshot.Total)
		require.Equal(t, 0, snapshot.Completed)
		require.Equal(t, 1, snapshot.NotStarted)
		require.Equal(t, 1, snapshot.InProgress)
		require.Equal(t, float64(25), snapshot.Percent)
	}

	// mark the first step as done
	{
		time.Sleep(200 * time.Millisecond)
		step1 := prog.Get("step1")
		step1.Done()
		require.Equal(t, progress.StateDone, step1.State)

		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 2)

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateInProgress, snapshot.State)
		require.Equal(t, 2, snapshot.Total)
		require.Equal(t, 1, snapshot.Completed)
		require.Equal(t, 1, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.Equal(t, float64(50), snapshot.Percent)
	}

	// mark the second step as done without starting it first
	{
		step2 := prog.Get("step2")
		step2.Done()
		require.Equal(t, progress.StateDone, step2.State)

		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 2)

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateDone, snapshot.State)
		require.Equal(t, 2, snapshot.Total)
		require.Equal(t, 2, snapshot.Completed)
		require.Equal(t, 0, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.Equal(t, float64(100), snapshot.Percent)
	}

	// add a third step
	{
		prog.AddStep("step3")
		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 3)
		require.NotNil(t, prog.Get("step3"))

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateInProgress, snapshot.State)
		require.Equal(t, 3, snapshot.Total)
		require.Equal(t, 2, snapshot.Completed)
		require.Equal(t, 1, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.Equal(t, 66, int(snapshot.Percent))
	}

	// add a fourth step
	{
		prog.AddStep("step4")
		require.NotEmpty(t, prog.Steps)
		require.Len(t, prog.Steps, 4)
		require.NotNil(t, prog.Get("step4"))

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateInProgress, snapshot.State)
		require.Equal(t, 4, snapshot.Total)
		require.Equal(t, 2, snapshot.Completed)
		require.Equal(t, 2, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.Equal(t, float64(50), snapshot.Percent)
	}

	// start step3 and step4 at the same time
	{
		step3 := prog.Get("step3")
		step4 := prog.Get("step4")
		step3.Start()
		step4.Start()

		snapshot := prog.Snapshot()
		require.Equal(t, "step3, step4", snapshot.Doing)
		require.Equal(t, progress.StateInProgress, snapshot.State)
		require.Equal(t, 4, snapshot.Total)
		require.Equal(t, 2, snapshot.Completed)
		require.Equal(t, 0, snapshot.NotStarted)
		require.Equal(t, 2, snapshot.InProgress)
		require.Equal(t, float64(75), snapshot.Percent)
	}

	// mark step3 and step4 as done at the same time
	{
		time.Sleep(200 * time.Millisecond)
		step1 := prog.Get("step1")
		step2 := prog.Get("step2")
		step3 := prog.Get("step3")
		step4 := prog.Get("step4")
		step3.Done()
		step4.Done()

		snapshot := prog.Snapshot()
		require.Equal(t, "", snapshot.Doing)
		require.Equal(t, progress.StateDone, snapshot.State)
		require.Equal(t, 4, snapshot.Total)
		require.Equal(t, 4, snapshot.Completed)
		require.Equal(t, 0, snapshot.NotStarted)
		require.Equal(t, 0, snapshot.InProgress)
		require.Equal(t, float64(100), snapshot.Percent)

		require.True(t, step1.Duration() > 200*time.Millisecond && step1.Duration() < 400*time.Millisecond)
		require.Zero(t, step2.Duration())
		require.True(t, step3.Duration() > 200*time.Millisecond && step3.Duration() < 400*time.Millisecond)
		require.True(t, step4.Duration() > 200*time.Millisecond && step4.Duration() < 400*time.Millisecond)
		require.True(t, snapshot.TotalDuration > 400*time.Millisecond && snapshot.TotalDuration < 600*time.Millisecond)
	}

	// debug
	// fmt.Println(u.PrettyJSON(prog))
}